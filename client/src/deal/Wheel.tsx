import React, { useEffect, useRef } from 'react';
import Konva from 'konva';

interface WheelProps {
  game: any;
  onStop: () => void;
  onRespin?: () => void;
}

const Wheel: React.FC<WheelProps> = ({ game, onStop, onRespin }) => {
  const container = useRef<HTMLDivElement>(null);
  const spinning = game.state.value === 'wheel_spin';

  useEffect(() => {
    const content = container.current;
    if (!content) return;

    const rect = content.getBoundingClientRect();
    const width = Math.max(rect.width, 320);
    const height = Math.max(rect.height, 320);
    const stage = new Konva.Stage({ container: content, width, height });
    const layer = new Konva.Layer();

    const available = (game.presents as any[]).map((p, i) => ({ p, i })).filter(({ p }) => !p.given);
    const sectorAngleDeg = 360 / Math.max(available.length, 1);
    const radius = Math.min(width, height) / 2 - 20;
    const centerX = width / 2;
    const centerY = height / 2;

    const wheelGroup = new Konva.Group({ x: centerX, y: centerY });

    available.forEach(({ p }, k) => {
      const start = -90 + k * sectorAngleDeg;
      const mid = start + sectorAngleDeg / 2;
      const wedge = new Konva.Wedge({
        x: 0,
        y: 0,
        radius,
        angle: sectorAngleDeg,
        rotation: start,
        fill: k % 2 === 0 ? '#535456' : '#333435',
        stroke: '#707072',
        strokeWidth: 2,
      });
      wheelGroup.add(wedge);

      const label = new Konva.Text({
        x: 0,
        y: 0,
        text: String(p.content),
        fontSize: 18,
        fill: '#ffffff',
        align: 'center',
        verticalAlign: 'middle',
      });
      const midRad = (mid * Math.PI) / 180;
      label.position({
        x: (radius * 0.7) * Math.cos(midRad) - label.width() / 2,
        y: (radius * 0.7) * Math.sin(midRad) - label.height() / 2,
      });
      wheelGroup.add(label);
    });

    const pointer = new Konva.Line({
      points: [
        centerX - 14, centerY - radius - 6,
        centerX + 14, centerY - radius - 6,
        centerX, centerY - radius + 14,
      ],
      closed: true,
      fill: '#ffffff',
    });
    layer.add(wheelGroup);
    layer.add(pointer);

    stage.add(layer);
    layer.draw();

    let animation: Konva.Animation | null = null;
    if (spinning) {
      const resultIndex = game.state.wheelResultIndex;
      const resultSector = available.findIndex(({ i }) => i === resultIndex);
      const resultDeg = -(resultSector + 0.5) * sectorAngleDeg;

      const cheat = game.state.cheatIndex >= 0;
      const laps = 5;

      // Phase 1 target: with a cheat we first stop near a non-result sector,
      // then respin to the actual result.
      const firstSector = cheat ? (resultSector === 0 ? 1 : 0) : resultSector;
      const firstDeg = -(firstSector + 0.5) * sectorAngleDeg;
      const firstAngle = laps * 360 + firstDeg;

      const respin = 360 + (resultDeg - firstDeg);

      const duration = 20000;
      const firstDuration = cheat ? 16000 : duration;
      const secondDuration = duration - firstDuration;
      const start = Date.now();

      animation = new Konva.Animation(() => {
        const elapsed = Date.now() - start;
        if (!cheat) {
          const progress = Math.min(elapsed / duration, 1);
          const eased = 1 - Math.pow(1 - progress, 3);
          wheelGroup.rotation(firstAngle * eased);
          if (progress >= 1) {
            animation?.stop();
            onStop();
          }
          return;
        }

        const phase = elapsed < firstDuration ? 1 : 2;
        if (phase === 1) {
          const progress = Math.min(elapsed / firstDuration, 1);
          const eased = 1 - Math.pow(1 - progress, 3);
          wheelGroup.rotation(firstAngle * eased);
          if (progress >= 1) {
            onRespin?.();
          }
        } else {
          // Respin to the cheat target.
          const progress = Math.min((elapsed - firstDuration) / secondDuration, 1);
          const eased = 1 - Math.pow(1 - progress, 3);
          wheelGroup.rotation(firstAngle + respin * eased);
          if (progress >= 1) {
            animation?.stop();
            onStop();
          }
        }
      }, layer);
      animation.start();
    }

    return () => {
      animation?.stop();
      stage.destroy();
    };
  }, [game.state.value, game.state.wheelResultIndex, game.state.cheatIndex, game.presents, spinning, onStop, onRespin]);

  return <div style={{ width: '100%', height: '100%', minHeight: 320 }} ref={container} />;
};

export default Wheel;
