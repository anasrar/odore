// pl26.ppk
// vi:ft=c

struct Bone {
    float x [[color("dc2626")]];
    float y [[color("65a30d")]];
    float z [[color("2563eb")]];
    s16 parent_bone [[color("7c3aed")]];
    s16 unknown0 [[color("525252")]];
};

struct VerterxBufferHeader {
    u32 position_address [[color("4f46e5")]];
    u32 normal_address [[color("3730a3")]];
    u32 uv_address [[color("2563eb")]];
    u32 color_address [[color("1e40af")]];
    u32 skin_address [[color("7c3aed")]];
    u16 vertex_total [[color("ca8a04")]];
    u8 unknown0 [[color("c026d3")]];
    u8 material_index [[color("ca8a04")]];
};

struct MDBHeader {
    u32 signature [[color("dc2626")]];
    u32 bone_address [[color("4f46e5")]];
    u16 bone_total [[color("0284c7")]];
    u16 vertex_buffer_total [[color("0284c7")]];
    u32 unknown0 [[color("c026d3")]];
    u32 skinning_format [[color("65a30d")]];
    u32 bone_index_table_address [[color("4f46e5")]];
    u32 bone_palette_table_address [[color("4f46e5")]];
    float scale [[color("c026d3")]];
    u32 vertex_buffer_addresses[vertex_buffer_total] [[color("4f46e5")]];
};

MDBHeader mdbheader_at_0x2EB0 @ 0x2EB0;
VerterxBufferHeader verterxbufferheader_at_0x3130 @ 0x3130;
Bone bone_at_0x2F30[mdbheader_at_0x2EB0.bone_total] @ 0x2F30;
