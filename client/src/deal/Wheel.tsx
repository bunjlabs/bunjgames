import React, { useEffect, useRef } from 'react';
import Konva from 'konva';

interface WheelProps {
  game: any;
  onStop: () => void;
}

const Wheel: React.FC<WheelProps> = ({ game, onStop }) => {
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
    const sectorAngle = (2.0 * Math.PI) / Math.max(available.length, 1);
    const radius = Math.min(width, height) / 2 - 20;
    const centerX = width / 2;
    const centerY = height / 2;

    available.forEach(({ i }, k) => {
      const angle = sectorAngle * k;
      const wedge = new Konva.Wedge({
        x: centerX,
        y: centerY,
        radius,
        angle: (sectorAngle * 180) / Math.PI,
        rotation: (angle * 180) / Math.PI,
        fill: k % 2 === 0 ? 'var(--bg-button)' : '#535456',
        stroke: 'var(--bg-dark)',
        strokeWidth: 2,
      });
      layer.add(wedge);

      const label = new Konva.Text({
        x: centerX,
        y: centerY,
        text: String(i + 1),
        fontSize: 18,
        fill: 'var(--text)',
        align: 'center',
        verticalAlign: 'middle',
      });
      const mid = angle + sectorAngle / 2;
      label.position({
        x: centerX + (radius * 0.7) * Math.cos(mid) - label.width() / 2,
        y: centerY + (radius * 0.7) * Math.sin(mid) - label.height() / 2,
      });
      layer.add(label);
    });

    const pointer = new Konva.Line({
      points: [
        centerX, centerY - radius,
        centerX, centerY - radius + 20,
      ],
      stroke: 'red',
      strokeWidth: 6,
    });
    layer.add(pointer);

    stage.add(layer);
    layer.draw();

    let animation: Konva.Animation | null = null;
    if (spinning) {
      const resultIndex = game.state.wheelResultIndex;
      const resultSector = available.findIndex(({ i }) => i === resultIndex);
      const targetMid = (resultSector + 1.5) * sectorAngle;

      const laps = 5;
      const totalAngle = laps * 2 * Math.PI + targetMid;
      const duration = 4000;
      const start = Date.now();

      animation = new Konva.Animation(() => {
        const elapsed = Date.now() - start;
        const progress = Math.min(elapsed / duration, 1);
        const eased = 1 - Math.pow(1 - progress, 3);
        layer.rotation((totalAngle * eased * 180) / Math.PI);
        if (progress >= 1) {
          animation?.stop();
          onStop();
        }
      }, layer);
      animation.start();
    }

    return () => {
      animation?.stop();
      stage.destroy();
    };
  }, [game.state.value, game.state.wheelResultIndex, game.presents, spinning, onStop]);

  return <div style={{ width: '100%', height: '100%', minHeight: 320 }} ref={container} />;
};

export default Wheel;
