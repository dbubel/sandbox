const std = @import("std");

pub fn main() !void {
    // Create an allocator
    const allocator = std.heap.page_allocator;

    // Define the key and value types for the HashMap
    // Initialize the HashMap
    var m = std.AutoArrayHashMap(usize, std.ArrayList(@Vector(2, f32))).init(allocator);
    // var map = std.hash_map.HashMap(KeyType, ValueType).init(allocator);
    defer m.deinit();
    var arr = std.ArrayList(@Vector(2, f32)).init(allocator);
    try arr.append(@Vector(2, f32){ 1, 2 });
    try m.put(1, arr);

    // Insert some values into the HashMap
    // try m.put("key1", 100);
    // try m.put("key2", 200);
    // try m.put("key3", 300);
    //
    // // Retrieve and print values from the HashMap
    // const val1 = m.get("key1");
    // if (val1) |value| {
    //     std.debug.print("key1: {}\n", .{value});
    // } else {
    //     std.debug.print("key1 not found\n", .{});
    // }
    //
    // const val2 = m.get("key2");
    // if (val2) |value| {
    //     std.debug.print("key2: {}\n", .{value});
    // } else {
    //     std.debug.print("key2 not found\n", .{});
    // }
    //
    // // Check for a key that doesn't exist
    // const val4 = m.get("key4");
    // if (val4) |value| {
    //     std.debug.print("key4: {}\n", .{value});
    // } else {
    //     std.debug.print("key4 not found\n", .{});
    // }
    //
    // // Remove a key from the HashMap
    // const removed = m.remove("key2");
    // if (removed) |value| {
    //     std.debug.print("Removed key2: {}\n", .{value});
    // } else {
    //     std.debug.print("key2 not found for removal\n", .{});
    // }
    //
    // // Attempt to retrieve the removed key
    // const val2_after_removal = m.get("key2");
    // if (val2_after_removal) |value| {
    //     std.debug.print("key2 after removal: {}\n", .{value});
    // } else {
    //     std.debug.print("key2 not found after removal\n", .{});
    // }
}
