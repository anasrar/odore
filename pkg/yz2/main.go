package yz2

import (
	"bytes"
	"fmt"
	"strconv"
)

const (
	YZ2HeaderSize  = 0x20
	YZ2TrailerSize = 4
	MaxYZ2Size     = uint64(0xfffffffe)
)

type Header struct {
	CompressedSize   uint64
	DecompressedSize uint64
	DataOffset       int
}

type rangeDecoder struct {
	data     []byte
	position int
	interval uint32
	code     uint32
}

func newRangeDecoder(data []byte) (*rangeDecoder, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("the compressed stream is empty")
	}
	return &rangeDecoder{
		data:     data,
		position: 1,
		interval: 0x80,
		code:     uint32(data[0]),
	}, nil
}

func (d *rangeDecoder) readByte() (byte, error) {
	if d.position >= len(d.data) {
		return 0, fmt.Errorf("the range-coded stream is truncated")
	}
	value := d.data[d.position]
	d.position++
	return value, nil
}

func (d *rangeDecoder) normalize() error {
	if d.interval > 0x800000 {
		return nil
	}

	byteCount := 1
	if d.interval <= 0x8000 {
		byteCount = 2
		if d.interval <= 0x80 {
			byteCount = 3
		}
	}

	for i := 0; i < byteCount; i++ {
		value, err := d.readByte()
		if err != nil {
			return err
		}
		d.interval <<= 8
		d.code = (d.code << 8) | uint32(value)
	}
	return nil
}

type rangeEncoder struct {
	low       uint64
	interval  uint64
	cache     uint64
	pendingFF int
	output    []byte
}

func newRangeEncoder() *rangeEncoder {
	return &rangeEncoder{interval: uint64(1) << 32}
}

func (e *rangeEncoder) shiftLow() {
	if (e.low >> 24) != 0xff {
		carry := e.low >> 32
		e.output = append(e.output, byte((e.cache+carry)&0xff))
		for i := 0; i < e.pendingFF; i++ {
			e.output = append(e.output, byte((0xff+carry)&0xff))
		}
		e.pendingFF = 0
		e.cache = (e.low & 0xffffffff) >> 24
	} else {
		e.pendingFF++
	}

	e.low = ((e.low & 0xffffffff) << 8) & 0xffffffff
}

func (e *rangeEncoder) encode(model *adaptiveModel, symbol int) error {
	if symbol < 0 || symbol >= model.symbolCount {
		return fmt.Errorf(
			"symbol %#x is outside a %#x-symbol model",
			symbol,
			model.symbolCount,
		)
	}

	for e.interval <= 0x1000000 {
		e.shiftLow()
		e.interval <<= 8
	}

	record := model.records[symbol]
	unit := e.interval >> 15
	e.low += unit * uint64(record.cumulative)
	e.interval = (unit * uint64(record.frequency)) & 0xffffffff
	e.interval &= 0xfffffffe
	model.update(symbol)
	return nil
}

func (e *rangeEncoder) finish() ([]byte, error) {
	e.low += e.interval >> 1
	for e.interval <= 0x2000000 {
		e.shiftLow()
		e.interval <<= 8
	}
	for i := 0; i < 2; i++ {
		e.shiftLow()
	}

	if e.pendingFF != 0 {
		e.output = append(e.output, byte(e.cache))
	}
	if len(e.output) == 0 || e.output[0] != 0 {
		return nil, fmt.Errorf("the range encoder did not produce its expected dummy byte")
	}
	return append([]byte(nil), e.output[1:]...), nil
}

type probabilityRecord struct {
	frequency  uint32
	cumulative uint32
}

type adaptiveModel struct {
	symbolCount int
	records     []probabilityRecord
	counts      []uint32
	total       uint32
	scaleBits   uint
	nextRebuild uint32
	lookup      []int
}

func newAdaptiveModel(symbolCount int) *adaptiveModel {
	initial := make([]uint32, symbolCount)
	for i := 0; i < 0x8000; i++ {
		initial[i%symbolCount]++
	}

	counts := make([]uint32, symbolCount)
	for i := range counts {
		counts[i] = 1
	}

	var scaleBits uint
	if symbolCount != 0 {
		for {
			scaleBits++
			if scaleBits > 14 || (1<<scaleBits) > symbolCount {
				break
			}
		}
	}

	return &adaptiveModel{
		symbolCount: symbolCount,
		records:     makeProbabilityRecords(initial, 1),
		counts:      counts,
		total:       uint32(symbolCount),
		scaleBits:   scaleBits,
		nextRebuild: uint32(1 << scaleBits),
	}
}

func makeProbabilityRecords(counts []uint32, multiplier uint32) []probabilityRecord {
	records := make([]probabilityRecord, len(counts))
	var cumulative uint32
	for i, count := range counts {
		scaled := (count * multiplier) & 0xffff
		records[i] = probabilityRecord{
			frequency:  scaled,
			cumulative: cumulative,
		}
		cumulative = (cumulative + scaled) & 0xffff
	}
	return records
}

func (m *adaptiveModel) buildLookup() error {
	lookup := make([]int, 0, 0x8000)
	for symbol, record := range m.records {
		expected := int(record.cumulative)
		if len(lookup) != expected {
			return fmt.Errorf(
				"invalid probability table at symbol %d: expected cumulative %d, got %d",
				symbol,
				len(lookup),
				expected,
			)
		}
		for i := uint32(0); i < record.frequency; i++ {
			lookup = append(lookup, symbol)
		}
	}
	m.lookup = lookup
	return nil
}

func (m *adaptiveModel) decode(coder *rangeDecoder) (int, error) {
	if err := coder.normalize(); err != nil {
		return 0, err
	}

	unit := coder.interval >> 14
	if unit == 0 {
		return 0, fmt.Errorf("the range decoder reached a zero interval")
	}
	coder.interval = unit

	if m.lookup == nil {
		if err := m.buildLookup(); err != nil {
			return 0, err
		}
	}

	quotient := (coder.code / unit) & 0xffff
	if uint64(quotient) >= uint64(len(m.lookup)) {
		return 0, fmt.Errorf(
			"range quotient %#x exceeds the probability table (%#x)",
			quotient,
			len(m.lookup),
		)
	}

	symbol := m.lookup[quotient]
	record := m.records[symbol]
	coder.code -= unit * record.cumulative
	coder.interval = uint32((uint64(unit)*uint64(record.frequency))&0xffffffff) >> 1

	m.update(symbol)
	return symbol, nil
}

func (m *adaptiveModel) update(symbol int) {
	m.counts[symbol] = (m.counts[symbol] + 1) & 0xffff
	m.total++

	if m.scaleBits < 15 {
		if m.total == m.nextRebuild {
			multiplier := uint32(1 << (15 - m.scaleBits))
			m.records = makeProbabilityRecords(m.counts, multiplier)
			m.scaleBits++
			m.nextRebuild = uint32(1 << m.scaleBits)
			m.lookup = nil
		}
		return
	}

	if m.total > 0x7fff {
		m.records = makeProbabilityRecords(m.counts, 1)
		var total uint32
		for i, count := range m.counts {
			if count > 1 {
				count >>= 1
			}
			m.counts[i] = count
			total += count
		}
		m.total = total
		m.lookup = nil
	}
}

type history struct {
	writeIndex uint32
	positions  [512]uint32
	lengths    [512]uint32
}

type encoderHistory struct {
	writeIndex       int
	positions        [512]int
	lengths          [512]int
	firstBytes       [512]int
	slotsByFirstByte [256]map[int]struct{}
}

func newEncoderHistory() *encoderHistory {
	h := &encoderHistory{}
	for i := range h.firstBytes {
		h.firstBytes[i] = -1
	}
	return h
}

func (h *encoderHistory) add(data []byte, position, length int) {
	slot := h.writeIndex
	previousFirstByte := h.firstBytes[slot]
	if previousFirstByte >= 0 {
		delete(h.slotsByFirstByte[previousFirstByte], slot)
	}

	firstByte := int(data[position])
	h.positions[slot] = position
	h.lengths[slot] = length
	h.firstBytes[slot] = firstByte
	if h.slotsByFirstByte[firstByte] == nil {
		h.slotsByFirstByte[firstByte] = make(map[int]struct{})
	}
	h.slotsByFirstByte[firstByte][slot] = struct{}{}
	h.writeIndex = (slot + 1) & 0x1ff
}

func ParseHeader(data []byte) (Header, error) {
	if len(data) < YZ2HeaderSize {
		return Header{}, fmt.Errorf("the file is too small to contain a YZ2 header")
	}

	headerBytes := data[:YZ2HeaderSize]
	tab := bytes.IndexByte(headerBytes, '\t')
	newline := -1
	if tab >= 0 {
		if relative := bytes.IndexByte(headerBytes[tab+1:], '\n'); relative >= 0 {
			newline = tab + 1 + relative
		}
	}
	if tab <= 0 || newline <= tab+1 {
		return Header{}, fmt.Errorf("expected hexadecimal sizes separated by TAB and LF")
	}

	compressedSize, err := strconv.ParseUint(string(headerBytes[:tab]), 16, 64)
	if err != nil {
		return Header{}, fmt.Errorf("the header contains an invalid compressed size: %w", err)
	}
	decompressedSize, err := strconv.ParseUint(string(headerBytes[tab+1:newline]), 16, 64)
	if err != nil {
		return Header{}, fmt.Errorf("the header contains an invalid decompressed size: %w", err)
	}

	return Header{
		CompressedSize:   compressedSize,
		DecompressedSize: decompressedSize,
		DataOffset:       (newline + YZ2HeaderSize) &^ 0x1f,
	}, nil
}

func encodeLength(coder *rangeEncoder, model *adaptiveModel, length int) error {
	if length < 2 || uint64(length) > MaxYZ2Size {
		return fmt.Errorf("cannot encode dictionary length %d", length)
	}

	if length <= 0xfe {
		return coder.encode(model, length+1)
	}

	stored := uint64(length + 1)
	byteCount := 4
	lengthCode := 0
	if stored <= 0xffff {
		byteCount = 2
		lengthCode = 2
	} else if stored <= 0xffffff {
		byteCount = 3
		lengthCode = 1
	}
	if err := coder.encode(model, lengthCode); err != nil {
		return err
	}

	for shift := (byteCount - 1) * 8; shift >= 0; shift -= 8 {
		if err := coder.encode(model, int((stored>>shift)&0xff)); err != nil {
			return err
		}
	}
	return nil
}

func matchLength(data []byte, source, target int) int {
	maximum := len(data) - target
	length := 0
	for length < maximum && data[source+length] == data[target+length] {
		length++
	}
	return length
}

type dictionaryMatch struct {
	slot   int
	length int
	cached bool
}

func findMatch(data []byte, position int, h *encoderHistory) (dictionaryMatch, bool) {
	bestSlot := -1
	bestLength := 1
	bestAge := 512

	for slot := range h.slotsByFirstByte[data[position]] {
		source := h.positions[slot]
		if source >= position {
			continue
		}
		available := matchLength(data, source, position)
		if available < 2 {
			continue
		}

		age := (h.writeIndex - 1 - slot) & 0x1ff
		if available > bestLength || (available == bestLength && age < bestAge) {
			bestSlot = slot
			bestLength = available
			bestAge = age
		}
	}

	if bestSlot < 0 {
		return dictionaryMatch{}, false
	}
	cachedLength := h.lengths[bestSlot]
	return dictionaryMatch{
		slot:   bestSlot,
		length: bestLength,
		cached: cachedLength >= 2 && cachedLength == bestLength,
	}, true
}

func packStream(stream []byte, decompressedSize int) ([]byte, error) {
	header := []byte(fmt.Sprintf("%x\t%x\n", len(stream), decompressedSize))
	if len(header) > YZ2HeaderSize {
		return nil, fmt.Errorf("the hexadecimal size header does not fit in 32 bytes")
	}

	packed := make([]byte, YZ2HeaderSize+len(stream)+YZ2TrailerSize)
	copy(packed, header)
	copy(packed[YZ2HeaderSize:], stream)
	return packed, nil
}

func Compress(data []byte) ([]byte, int, error) {
	if len(data) == 0 {
		return nil, 0, fmt.Errorf("the game decoder cannot represent an empty output stream")
	}
	if uint64(len(data)) > MaxYZ2Size {
		return nil, 0, fmt.Errorf("the input is too large for YZ2's 32-bit lengths")
	}

	coder := newRangeEncoder()
	tokenModel := newAdaptiveModel(0x500)
	byteModel := newAdaptiveModel(0x100)
	var histories [256]*encoderHistory
	for i := range histories {
		histories[i] = newEncoderHistory()
	}

	position := 0
	tokenCount := 0
	for position < len(data) {
		var match dictionaryMatch
		var found bool
		if position != 0 {
			match, found = findMatch(data, position, histories[data[position-1]])
		}

		tokenLength := 1
		if !found {
			if err := coder.encode(tokenModel, 0x400+int(data[position])); err != nil {
				return nil, 0, err
			}
		} else {
			tokenLength = match.length
			h := histories[data[position-1]]
			relativeSlot := (match.slot - h.writeIndex) & 0x1ff
			if match.cached {
				if err := coder.encode(tokenModel, 0x200+relativeSlot); err != nil {
					return nil, 0, err
				}
			} else {
				if err := coder.encode(tokenModel, relativeSlot); err != nil {
					return nil, 0, err
				}
				if err := encodeLength(coder, byteModel, tokenLength); err != nil {
					return nil, 0, err
				}
			}
		}

		if position > 0 {
			histories[data[position-1]].add(data, position, tokenLength)
		}
		position += tokenLength
		tokenCount++
	}

	stream, err := coder.finish()
	if err != nil {
		return nil, 0, err
	}
	packed, err := packStream(stream, len(data))
	if err != nil {
		return nil, 0, err
	}

	if coder.pendingFF != 0 {
		verified, _, decodeErr := Decompress(packed)
		if decodeErr != nil || !bytes.Equal(verified, data) {
			for i := 0; i < coder.pendingFF; i++ {
				coder.output = append(coder.output, 0xff)
			}
			coder.pendingFF = 0
			stream = append([]byte(nil), coder.output[1:]...)
			packed, err = packStream(stream, len(data))
			if err != nil {
				return nil, 0, err
			}
		}
	}

	return packed, tokenCount, nil
}

func Decompress(data []byte) ([]byte, int, error) {
	header, err := ParseHeader(data)
	if err != nil {
		return nil, 0, err
	}
	if header.DecompressedSize > MaxYZ2Size {
		return nil, 0, fmt.Errorf(
			"the declared output size %#x exceeds YZ2's 32-bit limit",
			header.DecompressedSize,
		)
	}
	maxInt := uint64(^uint(0) >> 1)
	if header.DecompressedSize > maxInt {
		return nil, 0, fmt.Errorf(
			"the declared output size %#x does not fit this platform's int type",
			header.DecompressedSize,
		)
	}

	available := len(data) - header.DataOffset
	if available < YZ2TrailerSize || header.CompressedSize > uint64(available-YZ2TrailerSize) {
		return nil, 0, fmt.Errorf(
			"the file declares %d compressed bytes but is truncated",
			header.CompressedSize,
		)
	}
	streamEnd := header.DataOffset + int(header.CompressedSize) + YZ2TrailerSize
	coder, err := newRangeDecoder(data[header.DataOffset:streamEnd])
	if err != nil {
		return nil, 0, err
	}

	tokenModel := newAdaptiveModel(0x500)
	byteModel := newAdaptiveModel(0x100)
	var histories [256]history

	outputSize := int(header.DecompressedSize)
	output := make([]byte, 0, outputSize)
	anchor := 0

	for len(output) < outputSize {
		token, err := tokenModel.decode(coder)
		if err != nil {
			return nil, 0, err
		}

		var tokenLength uint32
		if token >= 0x400 {
			output = append(output, byte(token))
			tokenLength = 1
		} else {
			if anchor >= len(output) {
				return nil, 0, fmt.Errorf("a dictionary token appeared before the first literal")
			}

			h := &histories[output[anchor]]
			historyIndex := (h.writeIndex + uint32(token)) & 0x1ff
			source := h.positions[historyIndex]

			if token < 0x200 {
				lengthCode, err := byteModel.decode(coder)
				if err != nil {
					return nil, 0, err
				}
				if lengthCode >= 3 {
					tokenLength = uint32(lengthCode - 1)
				} else {
					for i := 0; i < 4-lengthCode; i++ {
						value, err := byteModel.decode(coder)
						if err != nil {
							return nil, 0, err
						}
						tokenLength = (tokenLength << 8) | uint32(value)
					}
					tokenLength--
				}
			} else {
				tokenLength = h.lengths[historyIndex]
			}

			if tokenLength == 0 {
				return nil, 0, fmt.Errorf("a dictionary token decoded to a zero length")
			}
			if uint64(source) >= uint64(len(output)) {
				return nil, 0, fmt.Errorf(
					"dictionary source %#x is outside the %#x-byte output",
					source,
					len(output),
				)
			}
			if uint64(tokenLength) > uint64(outputSize-len(output)) {
				return nil, 0, fmt.Errorf(
					"the final token overshot the declared output size (%d + %d > %d)",
					len(output),
					tokenLength,
					outputSize,
				)
			}

			for i := uint32(0); i < tokenLength; i++ {
				output = append(output, output[int(source+i)])
			}
		}

		if anchor < len(output)-1 {
			h := &histories[output[anchor]]
			slot := h.writeIndex
			h.positions[slot] = uint32(anchor + 1)
			h.lengths[slot] = tokenLength
			h.writeIndex = (slot + 1) & 0x1ff
			anchor = len(output) - 1
		}
	}

	if len(output) != outputSize {
		return nil, 0, fmt.Errorf(
			"the final token overshot the declared output size (%d != %d)",
			len(output),
			outputSize,
		)
	}
	return output, coder.position, nil
}
