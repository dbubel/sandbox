const std = @import("std");

const Server = struct {
    port: u16,
    pub fn init(port: u16) Server {
        return Server{ .port = port };
    }

    pub fn run(s: Server) !void {
        var gpa = std.heap.GeneralPurposeAllocator(.{}){};

        // var arena = std.heap.ArenaAllocator.init(gpa.allocator());
        const arena_alloc = gpa.allocator();
        defer _ = gpa.deinit();

        const addr = std.net.Address.parseIp4("127.0.0.1", s.port) catch |err| {
            std.debug.print("error parsing address{any}\n", .{err});
            return;
        };

        var server_base = addr.listen(.{}) catch |err| {
            std.debug.print("error listening on address {any}\n", .{err});
            return;
        };
        defer server_base.deinit();

        const num_threads = 12;
        const threads = try gpa.allocator().alloc(std.Thread, num_threads);

        for (threads) |*t| {
            t.* = try std.Thread.spawn(.{}, handlerFn, .{ &server_base, arena_alloc });
        }

        std.debug.print("server started...\n", .{});
        for (threads) |t| {
            t.join();
        }
    }
};

fn handlerFn(base_server: *std.net.Server, alloc: std.mem.Allocator) void {
    while (true) {
        // blocks until a client connects
        var conn = base_server.accept() catch |err| {
            std.debug.print("error accept {any}\n", .{err});
            return;
        };
        defer conn.stream.close();

        // buffer for the incoming request header
        var header_buf: [1024]u8 = undefined;

        // create an http server based off of the open connection
        var server = std.http.Server.init(conn, &header_buf);

        // receiveHead reads the request header into the buffer
        // that was provided to the init function for the server
        var req = server.receiveHead() catch |err| {
            std.debug.print("{any}\n", .{err});
            return;
        };

        // example of reading headers from the incoming request
        std.debug.print("Headers:\n", .{});
        var headers = req.iterateHeaders();
        while (headers.next()) |header| {
            std.debug.print("\t{s}:{s}\n", .{ header.name, header.value });
        }

<<<<<<< HEAD
        _ = std.json.stringify(.{ .fuck = "fuck" }, .{ .whitespace = .minified }, aids.writer()) catch unreachable;
        _ = aids.flush() catch unreachable;
        _ = aids.end() catch unreachable;
        //
        // _ = req.respond("Hello http!\n", .{ .status = .ok }) catch unreachable;
=======
        // example of reading the request body. create a reader and then
        // use the readAll function to read the request body into request_buf
        const reader = req.reader() catch |err| {
            std.debug.print("error getting request reader {any}\n", .{err});
            return;
        };
        var request_buf: [1024]u8 = undefined;
        const n = reader.readAll(&request_buf) catch |err| {
            std.debug.print("error reading request {any}\n", .{err});
            return;
        };
>>>>>>> 2f052a9730e1282668124d3b742a90c8c81dfb1d

        std.debug.print("request body len {d}\n", .{n});
        std.debug.print("request body {s}\n", .{request_buf[0..n]});

        // here we attempt to parse an incoming json request body to a person struct
        const person = struct {
            name: []const u8,
        };
        var parsed = std.json.parseFromSlice(person, alloc, request_buf[0..n], .{}) catch |err| {
            std.debug.print("error parsing request body to json struct {any}\n", .{err});
            return;
        };
        defer parsed.deinit();

        // send a streaming json response with a custom status code
        var resp_buf: [1024]u8 = undefined;
        var resp_streamer = req.respondStreaming(.{ .respond_options = .{ .status = .bad_request }, .send_buffer = &resp_buf });

        const payload = .{ .hello = parsed.value.name };
        std.json.stringify(payload, .{}, resp_streamer.writer()) catch |err| {
            std.debug.print("error stringify {any}\n", .{err});
            return;
        };

        // end must be called to finalize the connection end
        resp_streamer.end() catch |err| {
            std.debug.print("error ending streaming response {any}\n", .{err});
            return;
        };
    }
}

pub fn main() !void {
    const server = Server.init(8080);
    try server.run();
}
