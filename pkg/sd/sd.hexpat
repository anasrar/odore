// r020.sd
// vi:ft=c

struct SDHeader {
    u32 signature [[color("dc2626")]];
    u32 scd_address [[color("4f46e5")]];
    u32 collision_address [[color("4f46e5")]];
    u32 mdb_address [[color("4f46e5")]];
    u32 unknown_mdb_address [[color("4f46e5")]];
};

SDHeader header @ 0x00;
