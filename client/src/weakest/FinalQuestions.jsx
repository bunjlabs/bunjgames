import React from "react";
import styles from "weakest/FinalQuestions.scss";
import classNames from "classnames";


const FinalQuestions = ({game}) => {
    const players = game.players.filter(player => !player.is_weak);
    const questions_info = game.final_questions;

    const first_player = players[0]
    const second_player = players[1]

    const first_questions = questions_info.filter(
        (item, index) => index % 2 === 0 && (index < 10 || item.is_processed)
    )
    const second_questions = questions_info.filter(
        (item, index) => index % 2 !== 0 && ((index + 1) / 2) <= first_questions.length
    )

    return <div className={styles.finalQuestions}>
        <div className={styles.playerBlock}>
            <div className={styles.player}>{first_player.name}</div>
            <div className={styles.player}>{second_player.name}</div>
        </div>
        <div className={styles.questionsBlock}>
            {[first_questions, second_questions].map((questions, index) =>
                <div key={index} className={styles.questions}>
                    {questions.map((item, index) =>
                        <div key={index} className={classNames(
                            item.is_correct && styles.correctQuestion,
                            (!item.is_correct && item.is_processed) && styles.incorrectQuestion,
                            styles.question
                        )} />
                    )}
                </div>
            )}
        </div>
    </div>;
}

export default FinalQuestions;
