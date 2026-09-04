// vi:ft=c

struct RegionAddressSize {
    u32 address [[color("4f46e5")]];
    u32 size [[color("0284c7")]];
};

struct ResourceEntry {
    u32 address [[color("4f46e5")]];
    u32 size [[color("0284c7")]];
    u32 unknown0 [[color("c026d3")]];
};

struct PPKHeader {
    u32 signature [[color("dc2626")]];
    u8 unknown0[4] [[color("57534e")]];
    RegionAddressSize region0;
    RegionAddressSize region1;
    RegionAddressSize region2;
    RegionAddressSize region3;
    RegionAddressSize region4;
    RegionAddressSize region5;
    RegionAddressSize region6;
    u32 mdb_total [[color("0284c7")]];
    ResourceEntry mdb_entries[mdb_total];
    u8 pad0[600-(mdb_total*12)] [[color("57534e")]];
    u32 t32_total [[color("0284c7")]];
    ResourceEntry t32_entries[t32_total];
    u8 pad1[600-(t32_total*12)] [[color("57534e")]];
};

PPKHeader header @ 0x00;
