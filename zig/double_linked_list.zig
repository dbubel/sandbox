const std = @import("std");
pub fn main() void {
    const ll = std.DoublyLinkedList(u32){};
    std.debug.print("{any}", .{ll});
    // ll.append(1);
    // const a = ll.popFirst();
    // std.debug.print("{any}\n", .{a});
}
