
Math.clamp = (x, a, b) => x > a ? a : x < b ? b : x;

Math.randomRange = (min, max) => Math.random() * (max - min) + min;

Math.randomRangeInt = (min, max) => Math.round(Math.randomRange(min, max));

Math.angleNormalize = (a) => {
    const normalized = a % (2 * Math.PI);
    return normalized < 0 ? 2 * Math.PI + normalized : normalized
}

Math.normalize = (a, norm) => {
    const normalized = a % norm;
    return normalized < 0 ? norm + normalized : normalized
}

window.css = (...a) => a.join(" ");