const std = @import("std");
const print = std.debug.print;

pub fn main() !void {
    var gpa = std.heap.GeneralPurposeAllocator(.{}){};
    defer _ = gpa.deinit();
    const allocator = gpa.allocator();

    const json_str = "[1,2,3]";
    const parsed = try std.json.parseFromSlice([]u8, allocator, json_str, .{});
    defer parsed.deinit();
    std.debug.print("{any}", .{parsed.value});
}
