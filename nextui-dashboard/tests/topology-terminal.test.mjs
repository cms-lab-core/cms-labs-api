import assert from 'node:assert/strict';
import test from 'node:test';
import { fitTerminalRectangle, resizeTerminalRectangle } from '../components/topology/terminal/geometry.ts';

const initial = { x: 100, y: 100, width: 400, height: 300 };

const assertInside = (rectangle, bounds) => {
  assert.ok(rectangle.x >= 0 && rectangle.y >= 0);
  assert.ok(rectangle.width >= 0 && rectangle.height >= 0);
  assert.ok(rectangle.x + rectangle.width <= bounds.width);
  assert.ok(rectangle.y + rectangle.height <= bounds.height);
};

test('terminal moves back inside the topology when checks expand', () => {
  const before = { ...initial, x: 950, y: 200 };
  const bounds = { width: 900, height: 700 };
  const after = fitTerminalRectangle(before, bounds);
  assert.equal(after.x, 490);
  assert.equal(after.width, 400);
  assertInside(after, bounds);
});

test('narrow workspace shrinks terminal instead of covering checks', () => {
  const bounds = { width: 280, height: 180 };
  const rectangle = fitTerminalRectangle(initial, bounds);
  assert.deepEqual(rectangle, { x: 10, y: 10, width: 260, height: 160 });
  assertInside(rectangle, bounds);
});

test('dragging is clamped on every edge of the local workspace', () => {
  const bounds = { width: 900, height: 700 };
  for (const x of [-1000, 0, 100, 2000]) {
    for (const y of [-1000, 0, 100, 2000]) {
      assertInside(fitTerminalRectangle({ ...initial, x, y }, bounds), bounds);
    }
  }
});

test('resize stops at workspace edge without moving the terminal origin', () => {
  const bounds = { width: 900, height: 700 };
  const rectangle = resizeTerminalRectangle(initial, { width: 5000, height: 5000 }, bounds);
  assert.deepEqual(rectangle, { x: 100, y: 100, width: 790, height: 590 });
  assertInside(rectangle, bounds);
  const small = resizeTerminalRectangle(initial, { width: -5000, height: -5000 }, bounds);
  assert.deepEqual(small, { x: 100, y: 100, width: 300, height: 200 });
});

test('restoring workspace width does not discard saved terminal geometry', () => {
  const before = { ...initial, x: 650, width: 600 };
  fitTerminalRectangle(before, { width: 500, height: 500 });
  assert.deepEqual(fitTerminalRectangle(before, { width: 1400, height: 800 }), before);
});

test('empty and tiny viewports never produce negative sizes or overflow', () => {
  for (const bounds of [
    { width: 0, height: 0 },
    { width: 1, height: 1 },
    { width: 12, height: 20 }
  ]) {
    assertInside(fitTerminalRectangle(initial, bounds), bounds);
    assertInside(resizeTerminalRectangle(initial, { width: 5000, height: 5000 }, bounds), bounds);
  }
});
