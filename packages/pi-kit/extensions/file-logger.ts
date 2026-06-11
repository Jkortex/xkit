/**
 * FileLogger — 安全的文件日志记录器
 *
 * - 写入 ~/.pi/agent/safe-run.log
 * - 每行一个 JSON 对象，便于 log aggregator 消费
 * - 自动创建目录
 */

import * as fs from 'node:fs';
import * as path from 'node:path';
import * as os from 'node:os';

export class FileLogger {
  private logPath: string;
  private writeStream: fs.WriteStream | null = null;

  constructor(filename = 'safe-run.log') {
    const logDir = path.join(os.homedir(), '.pi', 'agent');
    // 确保目录存在
    if (!fs.existsSync(logDir)) {
      fs.mkdirSync(logDir, { recursive: true });
    }
    this.logPath = path.join(logDir, filename);
  }

  private ensureStream(): fs.WriteStream {
    if (!this.writeStream || this.writeStream.destroyed) {
      const dir = path.dirname(this.logPath);
      if (!fs.existsSync(dir)) {
        fs.mkdirSync(dir, { recursive: true });
      }
      this.writeStream = fs.createWriteStream(this.logPath, {
        flags: 'a', // append
        encoding: 'utf-8',
      });
      // 忽略 ENOENT（测试环境 cleanup 后可能发生）
      this.writeStream.on('error', (err: NodeJS.ErrnoException) => {
        if (err.code !== 'ENOENT') {
          console.error('[safe-run] write stream error:', err);
        }
      });
      // 优雅退出时关闭
      process.once('exit', () => this.close());
    }
    return this.writeStream;
  }

  log(level: 'info' | 'warn' | 'error', event: string, details: Record<string, unknown>): void {
    const entry = JSON.stringify({
      timestamp: new Date().toISOString(),
      level,
      logger: 'safe-run',
      event,
      details,
    }) + '\n';

    try {
      const stream = this.ensureStream();
      stream.write(entry);
    } catch {
      // 写入失败时 fallback 到 stderr，避免静默丢失
      console.error('[safe-run] log write failed, falling back to stderr:', entry);
    }
  }

  close(): void {
    if (this.writeStream) {
      this.writeStream.end();
      this.writeStream = null;
    }
  }
}
