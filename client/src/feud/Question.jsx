import React from "react";
import styles from "./Question.scss";
import {FaTimes} from "react-icons/fa";
import classNames from "classnames";

const Question = ({game, showHiddenAnswers, className, onSelect}) => {
    const answerer = game.answerer && game.players.find(t => t.id === game.answerer);
    const answers = [];
    let strikesContainersCount = 0;
    game.question.answers.forEach((answer, index) => {
        let isAnswerOpened = (game.state !== "final_questions") ? answer.is_opened : showHiddenAnswers && answer.is_final_answered
        answers.push(
            <div key={'answer_' + index}
                 onClick={() => !isAnswerOpened && onSelect(answer.id)}
                 className={classNames(
                    isAnswerOpened && styles.opened,
                    !isAnswerOpened && onSelect ? styles.active : null,
                    styles.cell, styles.answer
                 )}
            >
                {(isAnswerOpened || showHiddenAnswers) && answer.text}
            </div>
        );
        answers.push(
            <div
                key={'value_' + index}
                className={classNames(
                    isAnswerOpened && styles.opened, styles.cell, styles.value
                )}
            >
                {(isAnswerOpened || showHiddenAnswers) && answer.value}
            </div>
        );
        if (answerer && strikesContainersCount < 3) {
            strikesContainersCount++;
            answers.push(
                <div
                    key={'strike_' + index}
                    className={classNames(
                        answerer && answerer.strikes >= strikesContainersCount && styles.active,
                        styles.cell, styles.strike
                    )}
                >
                    <FaTimes />
                </div>
            );
        } else {
            answers.push(
                <div key={'strike_' + index} />
            );
        }
    });

    return <div className={classNames(className, styles.grid)}>
        <div className={classNames(styles.cell, styles.question)}>{game.question.text}</div>
        {answers}
    </div>
};

const FinalQuestions = ({game, className}) => {
    const answers = [];
    game.final_questions.forEach((question, index) => {
        answers.push(
            <div key={'question_' + index} className={classNames(styles.cell, styles.answer, styles.opened)}>
                {!question.is_processed && question.text}
            </div>
        );
        let answer = question.answers.length ? question.answers[0] : null;
        answers.push(
            <div key={'answer_' + index} className={classNames(styles.cell, styles.answer, styles.opened)}>
                {!question.is_processed && (answer ? answer.text : "-")}
            </div>
        );
        answers.push(
            <div key={'value_' + index} className={classNames(styles.cell, styles.value, styles.opened)}>
                {!question.is_processed && (answer ? answer.value : "0")}
            </div>
        );
    });

    return <div className={classNames(className, styles.finalQuestionsGrid)}>
        {answers}
    </div>
}

export {Question, FinalQuestions}
