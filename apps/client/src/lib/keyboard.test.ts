import { describe, expect, test } from 'vitest';
import { hasBang, nextIndex } from './keyboard';

describe('nextIndex', () => {
	test('moves through results and back to the search box', () => {
		expect(nextIndex('ArrowDown', -1, 3)).toBe(0);
		expect(nextIndex('j', 1, 3)).toBe(2);
		expect(nextIndex('ArrowDown', 2, 3)).toBe(2);
		expect(nextIndex('k', 0, 3)).toBe(-1);
		expect(nextIndex('End', 0, 3)).toBe(2);
		expect(nextIndex('x', 0, 3)).toBeNull();
		expect(nextIndex('ArrowDown', -1, 0)).toBeNull();
	});
});

describe('hasBang', () => {
	test('finds a bang anywhere', () => {
		expect(hasBang('!g cats')).toBe(true);
		expect(hasBang('cats !i')).toBe(true);
		expect(hasBang('cats ! dogs')).toBe(false);
	});
});
