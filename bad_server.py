from http.server import HTTPServer, BaseHTTPRequestHandler

class BadServer(BaseHTTPRequestHandler):
    def do_GET(self):
        # If the file is "admin" or "secret", give real content
        if self.path == "/admin" or self.path == "/admin.php":
            self.send_response(200)
            self.end_headers()
            self.wfile.write(b"This is the REAL admin page!")
        elif self.path == "/secret":
            self.send_response(200)
            self.end_headers()
            self.wfile.write(b"This is the REAL secret!")
        else:
            # For EVERYTHING else, return 200 OK (The Trap!)
            # But the content is the generic "Whoops" page
            self.send_response(200)
            self.end_headers()
            self.wfile.write(b"Whoops! Page not found (But I'm sending 200 OK hehe)")

print("Starting Bad Server on port 8000...")
HTTPServer(('0.0.0.0', 8000), BadServer).serve_forever()