#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
EtherNet/IP Server using cpppo library
Based on: https://github.com/pjkundert/cpppo

完整支持:
- Class1/Class3 连接
- 标签读写 (Read/Write Tag)
- Assembly 对象
- Identity 对象
- 完整的 CIP 协议栈
- 支持 pycomm3 工程级识别
"""

import logging
import threading
import time
from cpppo.server.enip import main
from cpppo import apidict

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# =============================================================================
# EtherNet/IP Server 主类
# =============================================================================
class EthernetIPServer:
    def __init__(self, ip_address='0.0.0.0', port=44818):
        self.ip_address = ip_address
        self.port = port
        self._running = False
        self._server_thread = None
        self._server_control = apidict(timeout=1.0)
    
    def _server_main(self):
        """服务器主循环"""
        try:
            # 定义工程级标签（ControlLogix 风格）
            # 格式: tag_name=TYPE[count]
            tags = [
                # 基础标签（用于协议验证器测试）
                "BoolTag=BOOL[1]",
                "SintTag=SINT[1]",
                "IntTag=INT[1]",
                "DintTag=DINT[1]",
                "LintTag=LINT[1]",
                "UsintTag=USINT[1]",
                "UintTag=UINT[1]",
                "UdintTag=UDINT[1]",
                "UlintTag=ULINT[1]",
                "RealTag=REAL[1]",
                "LrealTag=LREAL[1]",
                "StringTag=STRING[82]",
                
                # Global Tags - 全局标签
                "Global.BoolTag=BOOL[1]",
                "Global.SintTag=SINT[1]",
                "Global.IntTag=INT[1]",
                "Global.DintTag=DINT[1]",
                "Global.LintTag=LINT[1]",
                "Global.UsintTag=USINT[1]",
                "Global.UintTag=UINT[1]",
                "Global.UdintTag=UDINT[1]",
                "Global.UlintTag=ULINT[1]",
                "Global.RealTag=REAL[1]",
                "Global.LrealTag=LREAL[1]",
                "Global.StringTag=STRING[82]",
                "Global.ByteTag=USINT[1]",
                "Global.WordTag=UINT[1]",
                "Global.DwordTag=UDINT[1]",
                "Global.LwordTag=ULINT[1]",
                
                # Integer Array
                "Global.IntArray[0]=INT[1]",
                "Global.IntArray[1]=INT[1]",
                "Global.IntArray[2]=INT[1]",
                "Global.IntArray[3]=INT[1]",
                "Global.IntArray[4]=INT[1]",
                
                # Program Tags - 程序作用域标签
                "Program:MainProgram.BoolTag=BOOL[1]",
                "Program:MainProgram.SintTag=SINT[1]",
                "Program:MainProgram.IntTag=INT[1]",
                "Program:MainProgram.DintTag=DINT[1]",
                "Program:MainProgram.LintTag=LINT[1]",
                "Program:MainProgram.UsintTag=USINT[1]",
                "Program:MainProgram.UintTag=UINT[1]",
                "Program:MainProgram.UdintTag=UDINT[1]",
                "Program:MainProgram.UlintTag=ULINT[1]",
                "Program:MainProgram.RealTag=REAL[1]",
                "Program:MainProgram.LrealTag=LREAL[1]",
                "Program:MainProgram.StringTag=STRING[82]",
                
                # IO Tags (Produced/Consumed)
                "Produced.OutputData=UDINT[1]",
                "Consumed.InputData=UDINT[1]",
            ]
            
            # 构建命令行参数
            argv = [
                f"-a", f"{self.ip_address}:{self.port}",
                "-v",
            ] + tags
            
            logger.info(f"Starting EtherNet/IP Server (cpppo) with {len(tags)} tags")
            
            # 启动 cpppo 服务器
            main.main(
                argv=argv,
                server=dict(control=self._server_control)
            )
            
        except Exception as e:
            logger.error(f"Server error: {e}", exc_info=True)
    
    def start(self):
        """启动服务器"""
        self._running = True
        self._server_thread = threading.Thread(target=self._server_main, daemon=True)
        self._server_thread.start()
        
        # 等待服务器启动
        time.sleep(2)
    
    def stop(self):
        """停止服务器"""
        self._running = False
        if hasattr(self._server_control, 'done'):
            self._server_control.done = True
        
        if self._server_thread:
            self._server_thread.join(timeout=10)
        
        logger.info("EtherNet/IP Server stopped")

# =============================================================================
# 主函数
# =============================================================================
if __name__ == '__main__':
    # 创建并启动服务器
    server = EthernetIPServer(ip_address='0.0.0.0', port=44818)
    server.start()
    
    logger.info("EtherNet/IP Server (cpppo) is running on port 44818")
    logger.info("Press Ctrl+C to stop")
    
    try:
        while True:
            time.sleep(1)
    except KeyboardInterrupt:
        logger.info("Received shutdown signal")
        server.stop()
