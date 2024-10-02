const std = @import("std");
const Server = struct {
    port: u16,
    pub fn init(port: u16) Server {
        return Server{ .port = port };
    }
    pub fn run(s: Server) !void {
        const addr = std.net.Address.parseIp4("127.0.0.1", s.port) catch |err| {
            std.debug.print("{any}\n", .{err});
            return;
        };

        var base_server = addr.listen(.{ .kernel_backlog = 1024, .reuse_address = true }) catch |err| {
            std.debug.print("error listen {any}\n", .{err});
            return;
        };

        defer base_server.deinit();

        var gpa = std.heap.GeneralPurposeAllocator(.{}){};
        defer _ = gpa.deinit();

        const num_threads = 12; //try std.Thread.getCpuCount();
        const threads = try gpa.allocator().alloc(std.Thread, num_threads);

        for (0.., threads) |i, *t| {
            t.* = try std.Thread.spawn(.{}, handlerThread, .{ &base_server, i });
        }

        for (threads) |t| {
            t.join();
        }
    }
};

fn handlerThread(base_server: *std.net.Server, i: usize) void {
    while (true) {
        var buf: [1024]u8 = undefined;
        var conn = base_server.accept() catch |err| {
            std.debug.print("error accept {any}\n", .{err});
            return;
        };
        _ = i;
        defer conn.stream.close();
        var server = std.http.Server.init(conn, &buf);

        var req = server.receiveHead() catch |err| {
            std.debug.print("{any}\n", .{err});
            return;
        };

        req.respond("hello", .{}) catch |err| {
            std.debug.print("{any}\n", .{err});
            return;
        };
    }
}
pub fn main() !void {
    const server = Server.init(8080);
    try server.run();
}
