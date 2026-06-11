import type { StatsDTO } from '@/application/ports/dto/Stats';
import { httpClient } from '../http/FetchClient';
import { AppError, Result, failure, success } from '@/utils/result';
import type { BackendStatsDTO } from './dto/BackendMemoDTO';
import { transformStats } from './transform/stats-transform';

export class HttpStatsGateway {
  async getStats(): Promise<Result<StatsDTO>> {
    const result = await httpClient.request<BackendStatsDTO>('/stats');

    if (result.kind === 'failure') return result;

    try {
      return success(transformStats(result.value));
    } catch (error) {
      return failure(
        new AppError('SERVER_ERROR', '统计数据解析失败，格式不合法', error),
      );
    }
  }
}

export const statsGateway = new HttpStatsGateway();
