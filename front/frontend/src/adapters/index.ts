import type { ScraperPort } from '@/types/adapter';
import { compositeScraperAdapter } from './compositeScraperAdapter';

export const scraperPort: ScraperPort = compositeScraperAdapter;

export { compositeScraperAdapter };
