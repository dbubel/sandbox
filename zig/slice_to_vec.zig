const std = @import("std");

pub fn main() !void {
    const arr: [3]i32 = [3]i32{ 1, 2, 3 };
    const vec2: @Vector(2, f32) = arr[0..2].*;
    std.debug.print("{any}", .{vec2});
}
