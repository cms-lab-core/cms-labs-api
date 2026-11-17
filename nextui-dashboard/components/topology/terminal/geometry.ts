export interface TerminalBounds {
  width: number;
  height: number;
}

export interface TerminalRectangle extends TerminalBounds {
  x: number;
  y: number;
}

export const fitTerminalRectangle = (rectangle: TerminalRectangle, bounds: TerminalBounds): TerminalRectangle => {
  const padding = Math.min(10, bounds.width / 2, bounds.height / 2);
  const width = Math.min(Math.max(300, rectangle.width), Math.max(0, bounds.width - padding * 2));
  const height = Math.min(Math.max(200, rectangle.height), Math.max(0, bounds.height - padding * 2));
  return {
    width,
    height,
    x: Math.max(padding, Math.min(rectangle.x, bounds.width - padding - width)),
    y: Math.max(padding, Math.min(rectangle.y, bounds.height - padding - height))
  };
};

export const resizeTerminalRectangle = (
  rectangle: TerminalRectangle,
  delta: TerminalBounds,
  bounds: TerminalBounds
): TerminalRectangle => {
  const fitted = fitTerminalRectangle(rectangle, bounds);
  const padding = Math.min(10, bounds.width / 2, bounds.height / 2);
  return {
    ...fitted,
    width: Math.min(Math.max(300, fitted.width + delta.width), bounds.width - padding - fitted.x),
    height: Math.min(Math.max(200, fitted.height + delta.height), bounds.height - padding - fitted.y)
  };
};
