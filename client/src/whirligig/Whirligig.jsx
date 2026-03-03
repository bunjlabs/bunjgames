import React, {useEffect, useRef} from "react";
import Konva from "konva";
import styles from "whirligig/Whirligig.module.scss";

const Whirligig = ({game, callback}) => {
    const container = useRef(null);

    useEffect(() => {
        const content = container.current;
        let rect = content.getBoundingClientRect();

        const stage = new Konva.Stage({
            container: content,
            width: rect.width,
            height: rect.height
        });

        const layer = new Konva.Layer();

        const circle = new Konva.Circle({
            stroke: 'yellow',
            strokeWidth: 4
        });
        layer.add(circle);

        const sectorAngle = 2.0 * Math.PI / game.items.length;
        const bars = [];
        const arrows = new Map();
        const items = new Map();

        for (let i=0; i < game.items.length; i++) {
            const item = game.items[i];

            const bar = new Konva.Line({
                stroke: 'yellow',
                strokeWidth: 4
            });
            bars.push(bar);
            layer.add(bar);

            if (item.is_processed) {
                const arrow = new Konva.Arrow({
                    stroke: 'yellow',
                    strokeWidth: 6,
                    fill: "yellow"
                });
                arrows.set(i, arrow);
                layer.add(arrow);
            } else {
                const has_city = item.questions.length === 1 && Boolean(item.questions[0].author_city)
                const text = new Konva.Text({
                    align: "center",
                    verticalAlign: "middle",
                    fill: "white",
                    fontSize: has_city ? 14: 20,
                    fontStyle: has_city ? "normal" : "bold",
                    text: has_city ? item.questions[0].author_city : item.name,
                    wrap: "word"
                })
                items.set(i, text);
                layer.add(text);
            }
        }

        const arrow = new Konva.Line({
            stroke: "red",
            strokeWidth: 8
        });
        layer.add(arrow);

        const top = new Konva.Circle({
            fill: 'white'
        });
        layer.add(top);

        stage.add(layer);
        layer.draw();

        stage.size({width: rect.width, height: rect.height});

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
                centerX + radius * Math.sin(Math.PI + angle), centerY + radius * Math.cos(Math.PI + angle)
            ])
            b.listening(false);
        });

        items.forEach((i, index) => {
            const angle = sectorAngle / 2 + sectorAngle * index;

            i.position({
                x: centerX + radius * 0.85 * Math.sin(angle),
                y: centerY - radius * 0.85 * Math.cos(angle)
            })

            const feelsBadMan = sectorAngle * radius * 0.85 * 0.8;
            i.width(feelsBadMan);
            i.fontSize(i.fontSize() === 20 ? screen * 0.04 : screen * 0.034);
            i.rotation(angle * 180 / Math.PI);
            i.offsetX(i.width() / 2);
            i.offsetY(i.height() / 2);
            i.listening(false);
        });

        arrows.forEach((i, index) => {
            const angle = sectorAngle / 2 + sectorAngle * index;

            i.strokeWidth(screen * 0.01)
            const halfOfArrowMagicNumber = screen * 0.05
            i.points([
                centerX + radius * 0.85 * Math.sin(angle) - halfOfArrowMagicNumber * Math.cos(angle),
                centerY + radius * 0.85 * Math.cos(angle + Math.PI) + halfOfArrowMagicNumber * Math.sin(angle + Math.PI),
                centerX + radius * 0.85 * Math.sin(angle) + halfOfArrowMagicNumber * Math.cos(angle),
                centerY + radius * 0.85 * Math.cos(angle + Math.PI) - halfOfArrowMagicNumber * Math.sin(angle + Math.PI),
            ])
            i.listening(false);
        });

        const t = Math.randomRange(30, 43);
        const reqAngle = sectorAngle * game.cur_random_item_idx + Math.random() * 0.98 * sectorAngle;
        let numberOfLaps = Math.randomRangeInt(30, 45);
        let S = numberOfLaps * 2 * Math.PI + reqAngle;
        const v0 = S / (t - t / 2);
        const a = 2.0 * (S - v0 * t) / Math.pow(t, 2);
        let ti = 0.0;

        const setAngle = (whirligigAngle) => {
            arrow.points([
                centerX - radius * Math.sin(whirligigAngle) * 0.2, centerY - radius * Math.cos(whirligigAngle + Math.PI) * 0.2,
                centerX + radius * Math.sin(whirligigAngle) * 0.8, centerY + radius * Math.cos(whirligigAngle + Math.PI) * 0.8,
            ])
        }
        const onUpdate = (delta) => {
            ti += delta;

            if (ti >= t) {
                setAngle(reqAngle);
                callback();
                anim.stop();
            } else {
                setAngle(Math.angleNormalize(
                    v0 * ti + (a * Math.pow(ti, 2)) / 2.0
                ));
            }
        }

        const anim = new Konva.Animation((frame) => {
            onUpdate(frame.timeDiff / 1000);
        }, layer);

        anim.start();
        return () => {
            anim.stop();
        }
    }, [callback, game.cur_random_item_idx, game.items])

    return <div className={styles.whirligig} ref={container}/>
};

export default Whirligig;
