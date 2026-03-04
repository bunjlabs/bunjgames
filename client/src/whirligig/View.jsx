import React, {useEffect, useCallback, useMemo} from "react";
import {AudioPlayer, HowlWrapper, ImagePlayer, VideoPlayer, Loading, useGame, useAuth} from "common/Essentials";
import {AdminAuth} from "common/Auth";
import Whirligig from "whirligig/Whirligig";
import styles from "whirligig/View.module.scss";
import {Content, ExitButton, GameView, TextContent} from "common/View";
import {useNavigate} from "react-router-dom";
import {GiMusicalNotes} from "react-icons/gi";
import {WHIRLIGIG_API} from "../index";


const QuestionsEndMusic = {
    current: 0,
    music: [
        HowlWrapper('/sounds/whirligig/question_end_1.mp3'),
        HowlWrapper('/sounds/whirligig/question_end_2.mp3'),
        HowlWrapper('/sounds/whirligig/question_end_3.mp3'),
        HowlWrapper('/sounds/whirligig/question_end_4.mp3'),
        HowlWrapper('/sounds/whirligig/question_end_5.mp3'),
    ]
}

const Music = {
    start: HowlWrapper('/sounds/whirligig/start.mp3'),
    intro: HowlWrapper('/sounds/whirligig/intro.mp3'),
    questions: HowlWrapper('/sounds/whirligig/questions.mp3'),
    whirligig: HowlWrapper('/sounds/whirligig/whirligig.mp3'),
    end: HowlWrapper('/sounds/whirligig/end_defeat.mp3'),
    end_victory: HowlWrapper('/sounds/whirligig/end_victory.mp3'),
    black_box: HowlWrapper('/sounds/whirligig/black_box.mp3'),
}

const Sounds = {
    sig1: HowlWrapper('/sounds/whirligig/sig1.mp3'),
    sig2: HowlWrapper('/sounds/whirligig/sig2.mp3'),
    sig3: HowlWrapper('/sounds/whirligig/sig3.mp3'),
    gong: HowlWrapper('/sounds/whirligig/gong.mp3'),
}

const loadSounds = () => {
    QuestionsEndMusic.music.forEach(m => m.load());
    Object.values(Music).forEach(m => m.load());
    Object.values(Sounds).forEach(m => m.load());
}

const resetSounds = () => {
    QuestionsEndMusic.music.forEach(m => m.stop());
    Object.values(Music).forEach(m => m.stop());
};

const isQuestionAvailable = (game) => {
    const {cur_question} = game;
    return ["question_start", "question_discussion", "answer"].includes(game.state)
        && ["text", "image", "audio", "video"].some(v => cur_question[v]);
}

const isAnswerAvailable = (game) => {
    const {cur_question} = game;
    return ["right_answer"].includes(game.state)
        && ["answer_text", "answer_image", "answer_audio", "answer_video"].some(v => cur_question[v]);
}

const isWhirligigAvailable = (game) => {
    return ["question_whirligig"].includes(game.state);
}

const QuestionMessage = React.memo(({game, text, image, audio, video}) => {
    const videoKey = useMemo(() => video ? `${game.state}-${video}` : null, [game.state, video]);
    const audioKey = useMemo(() => audio ? `${game.state}-${audio}` : null, [game.state, audio]);
    
    return <div className={styles.media}>
        {text && !image && !video && <div><p>{text}</p></div>}
        {image && <ImagePlayer game={game} url={image}/>}
        {["question_start", "right_answer"].includes(game.state) && audio &&
        <div key={audioKey}><AudioPlayer controls playing={true} game={game} url={audio}/></div>}
        {["question_start", "right_answer"].includes(game.state) && video &&
        <VideoPlayer key={videoKey} controls playing={true} game={game} url={video}/>}
        {!text && !image && !video && audio && <div><p style={{fontSize: "150px"}}><GiMusicalNotes /></p></div>}
    </div>
});

const triggerTimerSound = (game, time) => {
    if (!game.cur_item) return;

    if (game.cur_item.type === "standard") {
        switch (time) {
            case 60:
                Sounds.sig1.play();
                break;
            case 10:
                Sounds.sig2.play();
                break;
            case 0:
                Sounds.sig3.play();
                break;
            default:
                break;
        }
    } else {
        switch (time) {
            case 20:
                Sounds.sig1.play();
                break;
            case 0:
                Sounds.sig3.play();
                break;
            default:
                break;
        }
    }
}

const stateContent = (game) => {
    const onWhirligigReady = () => Music.whirligig.stop();

    if (isWhirligigAvailable(game)) {
        return <Whirligig game={game} callback={onWhirligigReady}/>
    } else if (isQuestionAvailable(game)) {
        const {text, image, audio, video} = game.cur_question;
        return <QuestionMessage
            game={game} text={text} image={image} audio={audio} video={video}
        />
    } else if (isAnswerAvailable(game)) {
        const {answer_text, answer_image, answer_audio, answer_video} = game.cur_question;
        return <QuestionMessage
            game={game} text={answer_text} image={answer_image} audio={answer_audio} video={answer_video}
        />
    } else {
        return <TextContent className={styles.score}>{game.connoisseurs_score} : {game.viewers_score}</TextContent>;
    }
}

const WhirligigView = () => {
    const onStateChange = useCallback((game) => {
        resetSounds();
        switch (game.state) {
            case "start": Music.start.play(); break;
            case "intro": Music.intro.play(); break;
            case "questions": Music.questions.play(); break;
            case "question_whirligig": Music.whirligig.play(); break;
            case "question_end":
                QuestionsEndMusic.music[QuestionsEndMusic.current].play();
                QuestionsEndMusic.current = (QuestionsEndMusic.current + 1) % QuestionsEndMusic.music.length;
                break;
            case "end":
                Music.end.play();
                break;
            default:
                break;
        }
    }, []);

    const onIntercom = useCallback((message) => {
        switch (message) {
            case "gong":
                Sounds.gong.play();
                break;
            case "sound_stop":
                resetSounds();
                break;
            default:
                break;
        }
    }, []);

    const game = useGame(WHIRLIGIG_API, onStateChange, onIntercom);

    useEffect(() => {
        loadSounds();
        return () => {
            resetSounds();
        }
    }, []);

    useEffect(() => {
        if (!game) return;

        let timer;
        if (game.timer_time > 0) {
            triggerTimerSound(game, WHIRLIGIG_API.calcTime());
            timer = setInterval(() => {
                triggerTimerSound(game, WHIRLIGIG_API.calcTime());
            }, 1000);
        }

        return () => timer && clearInterval(timer);
    }, [game]);

    const [connected, setConnected] = useAuth(WHIRLIGIG_API);

    const navigate = useNavigate();
    const onLogout = () => {
        WHIRLIGIG_API.logout();
        navigate("/admin");
    };

    if (!connected) return <AdminAuth api={WHIRLIGIG_API} setConnected={setConnected}/>;
    if (!game) return <Loading/>;

    return <GameView>
        <ExitButton onClick={onLogout}/>
        <Content>{stateContent(game)}</Content>
    </GameView>
}

export default WhirligigView;
