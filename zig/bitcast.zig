const std = @import("std");
const testing = std.testing;

// Define a packed struct
const Person = struct {
    age: u8,
    height: u16,
    is_student: bool,
};

pub fn main() !void {
    // Create an original struct
    var original = Person{
        .age = 30,
        .height = 180,
        .is_student = false,
    };

    // Convert struct to bytes
    const bytes = std.mem.asBytes(&original);

    // Recreate the struct from bytes
    const reconstructed = std.mem.bytesToValue(Person, bytes);
    // const reconstructed = @as(Person, @bitCast(bytes.*));

    // Verify the reconstruction
    try testing.expectEqual(original.age, reconstructed.age);
    try testing.expectEqual(original.height, reconstructed.height);
    try testing.expectEqual(original.is_student, reconstructed.is_student);

    std.debug.print("Original: age={}, height={}, student={}\n", .{ original.age, original.height, original.is_student });
    std.debug.print("Reconstructed: age={}, height={}, student={}\n", .{ reconstructed.age, reconstructed.height, reconstructed.is_student });
}
