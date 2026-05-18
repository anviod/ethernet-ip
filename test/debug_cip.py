#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
调试 CIP 响应格式
"""

import socket
import struct

def build_encap(cmd, session, payload):
    return struct.pack('<HHIIQI',
        cmd,
        len(payload),
        session,
        0x00000000,
        0x0000000000000000,
        0x00000000
    ) + payload

def build_cpf(cip_data):
    return (
        struct.pack('<I', 0x00000000) +
        struct.pack('<H', 2) +
        struct.pack('<HH', 0x0000, 0) +
        struct.pack('<HH', 0x00B2, len(cip_data)) +
        cip_data
    )

def send_rr(sock, session, cip, timeout=0):
    cpf = build_cpf(cip)
    payload = struct.pack('<I', 0)
    payload += struct.pack('<H', timeout)
    payload += cpf[4:]
    pkt = build_encap(0x006F, session, payload)
    sock.send(pkt)
    return sock.recv(4096)

def register_session(sock):
    payload = struct.pack('<HH', 0x0001, 0x0000)
    pkt = build_encap(0x0065, 0, payload)
    sock.send(pkt)
    resp = sock.recv(1024)
    session = struct.unpack('<I', resp[4:8])[0]
    return session

def main():
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sock.settimeout(5.0)
    sock.connect(("127.0.0.1", 44818))
    
    session = register_session(sock)
    print(f"Session: 0x{session:08X}")
    
    # 测试不同的路径格式
    test_paths = [
        # 标准 CIP 路径: Class 2, Instance 1, Attribute 1
        bytes([0x0E, 0x03, 0x20, 0x02, 0x24, 0x01, 0x30, 0x01]),
        # Logix 风格路径
        bytes([0x0E, 0x01, 0x20, 0x02]),
        # Symbolic path for "BoolTag"
        bytes([0x4C, 0x05, 0x91, 0x07, 0x42, 0x6F, 0x6F, 0x6C, 0x54, 0x61, 0x67, 0x00]),
    ]
    
    for i, cip in enumerate(test_paths):
        try:
            print(f"\n=== Test {i+1} ===")
            print(f"CIP Request: {cip.hex()}")
            resp = send_rr(sock, session, cip)
            print(f"Response length: {len(resp)}")
            print(f"Response hex: {resp.hex()}")
            
            if len(resp) >= 12:
                encap_status = struct.unpack('<I', resp[8:12])[0]
                print(f"Encap Status: 0x{encap_status:08X}")
                
        except Exception as e:
            print(f"Error: {e}")
    
    sock.close()

if __name__ == "__main__":
    main()
