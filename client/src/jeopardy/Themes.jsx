import React from "react";
import styles from "./Themes.scss";
import {FaCat} from "react-icons/fa";
import {RiAuctionFill} from "react-icons/ri"
import classNames from "classnames";

const Theme = ({theme, onSelect, active = false}) => (
    <div className={classNames(active && styles.active, styles.theme)}
         onClick={() => !theme.is_removed && onSelect(theme.id)}
         title={theme.name + "\n" + (theme.comment ?? "")}>
        <div>{!theme.is_removed && theme.name}</div>
    </div>
);

const Question = ({question, onSelect}) => (
    <div className={classNames(!question.is_processed && styles.active, styles.question)}
         onClick={() => !question.is_processed && onSelect(question.id)}>
        <div>
            {!question.is_processed && onSelect && question.type === "bagcat" && <FaCat/>}
            {!question.is_processed && onSelect && question.type === "auction" && <RiAuctionFill/>}
            {!question.is_processed && question.value}
        </div>
    </div>
);

const ThemesList = ({game, onSelect, active = false}) => (
    <div className={styles.themesList}>
        {game.themes.map((theme, index) => <Theme onSelect={onSelect} active={active} key={index} theme={theme}/>)}
    </div>
);

const ThemesGrid = ({game}) => (
    <div className={styles.themesGrid}>
        {game.themes.slice(0, game.themes.length - (game.themes.length % 3)).map((theme, index) => <Theme key={index} theme={theme}/>)}
    </div>
);

const QuestionsGrid = ({game, selectedId, onSelect}) => {
    let items = [];

    const maxQuestions = Math.max(...game.themes.map(t => t.questions.length));

    game.themes.forEach((theme, themeIndex) => {
        items.push(<Theme key={themeIndex} theme={theme}/>);
        theme.questions.forEach((question, questionIndex) =>
            items.push(<Question
                onSelect={onSelect}
                key={themeIndex + "_" + questionIndex}
                question={question}
            />)
        );
    });

    return <div className={styles.questionsGrid} style={{gridTemplateColumns: `minmax(0, 5fr) repeat(${maxQuestions}, minmax(0, 2fr))`}}>
        {items}
    </div>
};

export {ThemesList, ThemesGrid, QuestionsGrid}
