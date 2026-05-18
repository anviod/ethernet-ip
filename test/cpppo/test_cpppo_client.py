#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Test client for cpppo EtherNet/IP server
"""

import sys
import time
from cpppo.server.enip.get_attribute import proxy_simple

def main():
    print("Testing cpppo EtherNet/IP Server...")
    
    try:
        # 创建客户端连接
        with proxy_simple(host='127.0.0.1', port=44818) as client:
            print("✓ Connected to server")
            
            # 测试读取标签
            tags = ['BoolTag', 'IntTag', 'RealTag', 'StringTag']
            
            for tag in tags:
                try:
                    result = list(client.read(tag))
                    print(f"✓ Read {tag}: {result}")
                except Exception as e:
                    print(f"✗ Failed to read {tag}: {e}")
            
            # 测试写入标签
            try:
                result = list(client.write([('IntTag', [12345])]))
                print(f"✓ Write result: {result}")
                result = list(client.read('IntTag'))
                print(f"✓ Read IntTag after write: {result}")
            except Exception as e:
                print(f"✗ Failed to write IntTag: {e}")
                
    except Exception as e:
        print(f"✗ Connection failed: {e}")
        return 1
    
    return 0

if __name__ == "__main__":
    sys.exit(main())
