import React, { useEffect, useRef } from 'react';
import Konva from 'konva';

interface WhirligigProps {
  game: any;
  callback: () => void;
}

const Whirligig: React.FC<WhirligigProps> = ({ game, callback }) => {
  const container = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const content = container.current!;
    const rect = content.getBoundingClientRect();

    const stage = new Konva.Stage({ container: content, width: rect.width, height: rect.height });
    const layer = new Konva.Layer();

    const circle = new Konva.Circle({ stroke: 'yellow', strokeWidth: 4 });
    layer.add(circle);

    const sectorAngle = (2.0 * Math.PI) / game.items.length;
    const bars: Konva.Line[] = [];
    const arrows = new Map<number, Konva.Arrow>();
    const items = new Map<number, Konva.Text>();

    for (let i = 0; i < game.items.length; i++) {
      const item = game.items[i];

      const bar = new Konva.Line({ stroke: 'yellow', strokeWidth: 4 });
      bars.push(bar);
      layer.add(bar);

      if (item.is_processed) {
        const arrow = new Konva.Arrow({ stroke: 'yellow', strokeWidth: 6, fill: 'yellow' });
        arrows.set(i, arrow);
        layer.add(arrow);
      } else {
        const isLongName = item.name.length > 3;
        const text = new Konva.Text({
          align: 'center',
          verticalAlign: 'middle',
          fill: 'white',
          fontSize: isLongName ? 14 : 20,
          fontStyle: isLongName ? 'normal' : 'bold',
          text: item.name,
          wrap: 'word',
        });
        items.set(i, text);
        layer.add(text);
      }
    }

    const arrow = new Konva.Line({ stroke: 'red', strokeWidth: 8 });
    layer.add(arrow);

    const top = new Konva.Circle({ fill: 'white' });
    layer.add(top);

    stage.add(layer);
    layer.draw();

    stage.size({ width: rect.width, height: rect.height });

    const screen = Math.min(stage.width(), stage.height());
    const radius = screen / 2 - screen * 0.065;
    const centerX = stage.width() / 2;
    const centerY = stage.height() / 2;

    circle.x(centerX);
    circle.y(centerY);
    circle.radius(radius);
    circle.listening(false);

    top.x(centerX);
    top.y(centerY);
    top.radius(radius * 0.1);
    top.listening(false);

    bars.forEach((b, index) => {
      const angle = sectorAngle * index;
      b.points([
        centerX, centerY,
        centerX + radius * Math.sin(Math.PI + angle),
        centerY + radius * Math.cos(Math.PI + angle),
      ]);
      b.listening(false);
    });

    items.forEach((t, index) => {
      const angle = sectorAngle / 2 + sectorAngle * index;
      t.position({
        x: centerX + radius * 0.85 * Math.sin(angle),
        y: centerY - radius * 0.85 * Math.cos(angle),
      });

      const w = sectorAngle * radius * 0.85 * 0.8;
      t.width(w);
      t.fontSize(t.fontSize() === 20 ? screen * 0.04 : screen * 0.034);
      t.rotation((angle * 180) / Math.PI);
      t.offsetX(t.width() / 2);
      t.offsetY(t.height() / 2);
      t.listening(false);
    });

    arrows.forEach((a, index) => {
      const angle = sectorAngle / 2 + sectorAngle * index;
      a.strokeWidth(screen * 0.01);
      const half = screen * 0.05;
      a.points([
        centerX + radius * 0.85 * Math.sin(angle) - half * Math.cos(angle),
        centerY + radius * 0.85 * Math.cos(angle + Math.PI) + half * Math.sin(angle + Math.PI),
        centerX + radius * 0.85 * Math.sin(angle) + half * Math.cos(angle),
        centerY + radius * 0.85 * Math.cos(angle + Math.PI) - half * Math.sin(angle + Math.PI),
      ]);
      a.listening(false);
    });

    const t = Math.randomRange(30, 43);
    const reqAngle = sectorAngle * game.state.whirligigPosition + Math.random() * 0.98 * sectorAngle;
    const numberOfLaps = Math.randomRangeInt(30, 45);
    const totalS = numberOfLaps * 2 * Math.PI + reqAngle;
    const v0 = totalS / (t - t / 2);
    const a2 = (2.0 * (totalS - v0 * t)) / Math.pow(t, 2);
    let ti = 0.0;

    const setAngle = (whirligigAngle: number) => {
      arrow.points([
        centerX - radius * Math.sin(whirligigAngle) * 0.2,
        centerY - radius * Math.cos(whirligigAngle + Math.PI) * 0.2,
        centerX + radius * Math.sin(whirligigAngle) * 0.8,
        centerY + radius * Math.cos(whirligigAngle + Math.PI) * 0.8,
      ]);
    };

    const anim = new Konva.Animation((frame: any) => {
      ti += frame!.timeDiff / 1000;
      if (ti >= t) {
        setAngle(reqAngle);
        callback();
        anim.stop();
      } else {
        setAngle(Math.angleNormalize(v0 * ti + (a2 * Math.pow(ti, 2)) / 2.0));
      }
    }, layer);

    anim.start();
    return () => anim.stop();
  }, [callback, game.state.whirligigPosition, game.items]);

  return <div style={{ width: '100%', height: '100%' }} ref={container} />;
};

export default Whirligig;
