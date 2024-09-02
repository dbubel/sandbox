const std = @import("std");

const Vector = struct {
    x: f64,
    y: f64,
};

fn processVectors(vectors: []Vector) void {
    for (vectors) |*vector| {
        std.debug.print("Original Vector: ({}, {})\n", .{ vector.x, vector.y });
        // Modify the vector
        vector.x += 1.0;
        vector.y += 1.0;
        std.debug.print("Modified Vector: ({}, {})\n", .{ vector.x, vector.y });
    }
}

pub fn main() void {
    const a = @Vector(2, i32){ 0, 0 };
    const b = @Vector(2, i32){ 1, 1 };
    const dist = @as(f32, @reduce(.Add, (a - b) * (a - b)));
    const dist_sqrt = std.math.sqrt(dist);

    std.debug.print("dist: {d}\n", .{dist_sqrt});
}
