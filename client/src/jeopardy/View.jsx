import React, {useEffect, useState} from "react";

import {AudioPlayer, HowlWrapper, ImagePlayer, VideoPlayer, Loading, useGame, useAuth} from "../common/Essentials";
import {AdminAuth} from "../common/Auth";

import {ThemesList, ThemesGrid, QuestionsGrid} from "jeopardy/Themes";
import {getRoundName, EventType} from "jeopardy/Common";
import {Content, ExitButton, GameView, TextContent, QRCodeContent} from "common/View";
import styles from "jeopardy/View.module.scss";
import {useNavigate} from "react-router-dom";
import {generateClientUrl} from "../common/View";
import {GiMusicalNotes} from "react-icons/gi";
import {JEOPARDY_API} from "../index";


const Music = {
    intro: HowlWrapper('/sounds/jeopardy/intro.mp3'),
    themes: HowlWrapper('/sounds/jeopardy/themes.mp3'),
    round: HowlWrapper('/sounds/jeopardy/round.mp3'),
    minute: HowlWrapper('/sounds/jeopardy/minute.mp3'),
    auction: HowlWrapper('/sounds/jeopardy/auction.mp3'),
    bagcat: HowlWrapper('/sounds/jeopardy/bagcat.mp3'),
    game_end: HowlWrapper('/sounds/jeopardy/game_end.mp3'),
}

const Sounds = {
    skip: HowlWrapper('/sounds/jeopardy/skip.mp3'),
}


const loadSounds = () => {
    Object.values(Music).forEach(m => m.load());
    Object.values(Sounds).forEach(m => m.load());
}

const resetSounds = () => {
    Object.values(Music).forEach(m => m.stop());
};


const delocalise = (url) => {
    let result = url;
    if (url.slice(7).includes("https:")) {
        result = url.replace(window.location.protocol + "//" + window.location.hostname, "");
    }
    return result;
}

const QuestionMessage = ({game, text, image, audio, video, isContentPlaying}) => {

    return <div className={styles.media}>
        {text && !image && !video && <p>{text}</p>}
        {image && <ImagePlayer game={game} url={delocalise(image)}/>}
        {audio && <AudioPlayer controls playing={isContentPlaying} game={game} url={delocalise(audio)}/>}
        {video && <VideoPlayer controls playing={isContentPlaying} game={game} url={delocalise(video)}/>}
        {!text && !image && !video && audio && <p style={{fontSize: "150px"}}><GiMusicalNotes/></p>}
    </div>
}

const stateContent = (game, isContentPlaying) => {
    const {question} = game;
    const answerer = game.answerer && game.players.find(p => p.id === game.answerer);

    switch (game.state) {
        case "waiting_for_players":
            return <QRCodeContent value={generateClientUrl('/jeopardy/client?token=' + game.token)}>
                {game.token}
            </QRCodeContent>;
        case "themes_all":
            return <ThemesGrid game={game}/>;
        case "round":
            return <TextContent>{getRoundName(game)}</TextContent>;
        case "round_themes":
        case "final_themes":
            return <ThemesList game={game}/>;
        case "questions":
            return <QuestionsGrid game={game}/>;
        case "question_event":
            return <TextContent><EventType type={game.question.type}/></TextContent>;
        case "question": case "answer": case "final_question": case "final_answer":
            return <QuestionMessage
                game={game} text={question.text} image={question.image} audio={question.audio} video={question.video}
                isContentPlaying={isContentPlaying}
            />;
        case "question_end":
            let answer_image = question.answer_image;
            if (!question.answer_image && !question.answer_text && !question.answer_video) {
                answer_image = question.image;
            }
            return <QuestionMessage
                game={game} text={question.answer_text} image={answer_image}
                audio={question.answer_audio} video={question.answer_video}
                isContentPlaying={isContentPlaying}
            />;
        case "final_player_answer":
            return <TextContent>{answerer.final_answer || "⸻"}</TextContent>;
        case "final_player_bet":
            return <TextContent>{answerer.final_bet}</TextContent>;
        default:
            return <TextContent>Jeopardy</TextContent>;
    }
};

const JeopardyView = () => {
    const [isContentPlaying, setContentPlaying] = useState(true);
    const game = useGame(JEOPARDY_API, (game) => {
        resetSounds();
        if (["question", "question_end", "final_question"].includes(game.state) && !isContentPlaying) {
            setContentPlaying(true);
        }
        switch (game.state) {
            case "intro": Music.intro.play(); break;
            case "round": Music.round.play(); break;
            case "round_themes": Music.themes.play(); break;
            case "question_event":
                if (game.question.type === "auction") {
                    Music.auction.play();
                } else if (game.question.type === "bagcat") {
                    Music.bagcat.play();
                }
                break;
            case "final_answer": Music.minute.play(); break;
            case "game_end": Music.game_end.play(); break;
            default: break;
        }
    }, (message) => {
        switch (message) {
            case "skip": Sounds.skip.play(); break;
            case "sound_stop": resetSounds(); setContentPlaying(false); break;
            case "replay": setContentPlaying(true); break;
            default: break;
        }
    });

    useEffect(() => {
        loadSounds();
        return () => {
            resetSounds();
        }
    }, []);

    const [connected, setConnected] = useAuth(JEOPARDY_API);

    const navigate = useNavigate();
    const onLogout = () => {
        JEOPARDY_API.logout();
        navigate("/admin");
    };

    if (!connected) return <AdminAuth api={JEOPARDY_API} setConnected={setConnected}/>;
    if (!game) return <Loading/>;

    return <GameView>
        <ExitButton onClick={onLogout}/>
        <Content>{stateContent(game, isContentPlaying)}</Content>
    </GameView>
}

export default JeopardyView;
