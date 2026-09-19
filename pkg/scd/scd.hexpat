// r020.sd
// vi:ft=c

struct Vector3 {
    float x [[color("dc2626")]];
    float y [[color("16a34a")]];
    float z [[color("2563eb")]];
};

struct SceneObject {
    Vector3 position;
    Vector3 rotation;
    Vector3 scale;
    u32 index [[color("9333ea")]];
    u32 unknown0 [[color("525252")]];
    u32 unknown1 [[color("525252")]];
    u32 unknown2 [[color("525252")]];
    u32 unknown3 [[color("525252")]];
    u32 unknown4 [[color("525252")]];
    u32 unknown5 [[color("525252")]];
    u32 unknown6 [[color("525252")]];
    u32 unknown7 [[color("525252")]];
    u32 mdb_address [[color("0d9488")]];
    u32 unknown8 [[color("525252")]];
};

struct SCDHeader {
    u32 signature [[color("dc2626")]];
    u32 unknown0 [[color("525252")]];
    u32 unknown1 [[color("525252")]];
    u32 scene_object_total [[color("16a34a")]];
    u32 unknown2 [[color("525252")]];
    u32 unknown3 [[color("525252")]];
    u32 unknown4 [[color("525252")]];
    u32 unknown5 [[color("525252")]];
    SceneObject scene_objects[scene_object_total];
};

SCDHeader header @ 0xC0;
