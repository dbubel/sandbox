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

        var base_server = addr.listen(.{}) catch |err| {
            std.debug.print("{any}\n", .{err});
            return;
        };
        defer base_server.deinit();

        var gpa = std.heap.GeneralPurposeAllocator(.{}){};
        defer _ = gpa.deinit();

        const num_threads = 12; //try std.Thread.getCpuCount();
        const threads = try gpa.allocator().alloc(std.Thread, num_threads);

        // try server.listen(self.address);

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
        var buf: [500]u8 = undefined;
        var conn = base_server.accept() catch |err| {
            std.debug.print("{any}\n", .{err});
            return;
        };
        _ = i;
        // std.debug.print("handled on {d}\n", .{i});
        defer conn.stream.close();
        var server = std.http.Server.init(conn, &buf);

        var req = server.receiveHead() catch |err| {
            std.debug.print("{any}\n", .{err});
            return;
        };

        // std.time.sleep(std.time.ns_per_s);
        // std.debug.print("{any}\n", .{buf[0..req.head_end]});
        req.respond("hello", .{}) catch |err| {
            std.debug.print("{any}\n", .{err});
            return;
        };
        // req.server.connection.stream.close();
    }
}
pub fn main() !void {
    const server = Server.init(8080);
    try server.run();
}
