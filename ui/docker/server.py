#!/usr/bin/env python3
import http.server
import socketserver
import os

class CustomHTTPRequestHandler(http.server.SimpleHTTPRequestHandler):
    def do_GET(self):
        # Handle health check endpoint
        if self.path == '/health':
            self.send_response(200)
            self.send_header('Content-type', 'text/plain')
            self.end_headers()
            self.wfile.write(b'healthy\n')
            return
        
        # Serve static files
        return http.server.SimpleHTTPRequestHandler.do_GET(self)

if __name__ == "__main__":
    PORT = 3000
    
    with socketserver.TCPServer(("", PORT), CustomHTTPRequestHandler) as httpd:
        print(f"Server running on port {PORT}")
        httpd.serve_forever()
