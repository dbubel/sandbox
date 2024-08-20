const std = @import("std");
pub fn main() !void {
    var i: usize = 0;
    for (0..3) |_| {
        std.debug.print("{d}\n", .{i});
        defer i = i + 1;
    }
    std.debug.print("{d}\n", .{i});
}
