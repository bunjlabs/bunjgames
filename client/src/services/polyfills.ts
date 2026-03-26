declare global {
  interface Math {
    clamp(x: number, a: number, b: number): number;
    randomRange(min: number, max: number): number;
    randomRangeInt(min: number, max: number): number;
    angleNormalize(a: number): number;
    normalize(a: number, norm: number): number;
  }
}

Math.clamp = (x: number, a: number, b: number) =>
  x > a ? a : x < b ? b : x;

Math.randomRange = (min: number, max: number) =>
  Math.random() * (max - min) + min;

Math.randomRangeInt = (min: number, max: number) =>
  Math.round(Math.randomRange(min, max));

Math.angleNormalize = (a: number) => {
  const normalized = a % (2 * Math.PI);
  return normalized < 0 ? 2 * Math.PI + normalized : normalized;
};

Math.normalize = (a: number, norm: number) => {
  const normalized = a % norm;
  return normalized < 0 ? norm + normalized : normalized;
};

export {};
