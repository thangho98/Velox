/**
 * @velox/shared/hooks — Series hooks re-export
 *
 * Series hooks are exported from media/index.ts to keep media and series
 * hooks in the same module. This file re-exports them for backwards
 * compatibility with the barrel structure.
 */

export {
  useSeriesList,
  useSeriesDetail,
  useSeriesSearch,
  useSeasons,
  useEpisodes,
  useEpisode,
  seriesKeys,
} from './media/useSeries'
