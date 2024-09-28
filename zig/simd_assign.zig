const std = @import("std");

pub fn assignArrayElementsSIMD(comptime T: type, comptime N: usize, array: []T, value: T) void {
    const VecType = @Vector(N, T);
    const vecValue = @splat(VecType, value);

    // Ensure the array length is a multiple of N
    const length = array.len;
    const simd_length = length - (length % N);

    for (array[0..simd_length].chunks(@sizeOf(VecType))) |chunk| {
        const vecPtr = @ptrCast(*VecType, chunk.ptr);
        vecPtr.* = vecValue;
    }

    // Handle any remaining elements
    for (array[simd_length..]) |*elem| {
        elem.* = value;
    }
}

pub fn main() void {
    var allocator = std.heap.page_allocator;
    const array = try allocator.alloc(u32, 16);

    assignArrayElementsSIMD(u32, 4, array, 42);

    for (array) |elem| {
        std.debug.print("{} ", .{elem});
    }
    std.debug.print("\n", .{});

    allocator.free(array);
}
