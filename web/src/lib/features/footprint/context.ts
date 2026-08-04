import { createModelDataContext } from 'svatoms';
import type { FootprintOverview } from './types';

export const footprintCtx = createModelDataContext<FootprintOverview>({
	name: 'footprintCtx',
	initial: null
});
