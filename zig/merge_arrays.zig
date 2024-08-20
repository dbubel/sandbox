const std = @import("std");

pub fn main() !void {
    var array1 = [_]i32{ 1, 2, 3 };
    var array2 = [_]i32{ 4, 5, 6 };
    var slice = std.ArrayList(i32).init(std.heap.page_allocator);
    defer slice.deinit();

    try slice.appendSlice(&array1);
    try slice.appendSlice(&array2);
}
