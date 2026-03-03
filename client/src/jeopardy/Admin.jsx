import React, {useEffect, useState} from "react";
import {useNavigate} from "react-router-dom";

import {AdminAuth} from "common/Auth";
import {
    AudioPlayer,
    ImagePlayer,
    Loading,
    VideoPlayer,
    useGame,
    useAuth,
    OvalButton,
    ButtonLink, Button, Input, HorizontalList, TwoLineListItem
} from "common/Essentials"
import {BlockContent, Content, Footer, FooterItem, GameAdmin, Header, TextContent} from "common/Admin";

import {ThemesList, ThemesGrid, QuestionsGrid} from "jeopardy/Themes";
import {getStatusName, EventType, getRoundName} from "jeopardy/Common";
import styles from "jeopardy/Admin.scss";
import {FaVolumeMute} from "react-icons/fa";
import {MdReplayCircleFilled} from "react-icons/md"
import classNames from "classnames";
import {JEOPARDY_API} from "../index";


const QuestionEvent = ({question}) => {
    const {
        type, theme, custom_theme, value
    } = question;
    let themeDiv;
    if (custom_theme) {
        themeDiv = <div>Custom theme: {custom_theme}</div>;
    } else {
        themeDiv = <div>Theme: {theme}</div>;
    }
    return <BlockContent>
        <div>
            <div className={styles.type}><EventType type={type}/></div>
            {themeDiv}
            <div>Value: {value}</div>
        </div>
    </BlockContent>
}

const Question = ({game}) => {
    const {
        value, custom_theme, text, image, audio, video,
        answer, comment, answer_text, answer_image, answer_audio, answer_video
    } = game.question;

    return <BlockContent>
        <div>
            <div>Value: {value}</div>
            {custom_theme && <div>Custom theme: {custom_theme}</div>}
        </div>
        <div className={styles.media}>
            <div>{text && <p>{text}</p>}</div>
            <div>{image && <ImagePlayer controls={true} game={game} url={image}/>}</div>
            <div>{audio && <AudioPlayer controls={true} game={game} url={audio}/>}</div>
            <div>{video && <VideoPlayer controls={true} game={game} url={video}/>}</div>
        </div>
        <div className={styles.media}>
            <div>{answer}</div>
            {comment && <div>Comment: {comment}</div>}
            <div>{answer_text && <p>{answer_text}</p>}</div>
            <div>{answer_image && <ImagePlayer controls={true} game={game} url={answer_image}/>}</div>
            <div>{answer_audio && <AudioPlayer controls={true} game={game} url={answer_audio}/>}</div>
            <div>{answer_video && <VideoPlayer controls={true} game={game} url={answer_video}/>}</div>
        </div>
    </BlockContent>
}

const FinalBets = ({players, answerer}) => {
    const onClick = (playerId) => {
        JEOPARDY_API.intercom("do_bet:" + playerId)
    }

    return <div className={classNames(styles.padding, styles.players)}>
        {players.map((player, index) =>
            <Player key={index} balance={player.final_bet} selected={answerer && player.id === answerer}
                    name={player.name} onClick={() => onClick(player.id)}/>
        )}
    </div>
}

const FinalAnswers = ({players, answerer}) => {
    const onClick = (playerId) => {
        JEOPARDY_API.intercom("do_answer:" + playerId)
    }

    return <div className={classNames(styles.padding, styles.players)}>
        {players.map((player, index) =>
            <Player key={index} balance={player.final_answer || "⸻"} selected={answerer && player.id === answerer}
                    name={player.name} onClick={() => onClick(player.id)}/>
        )}
    </div>
}

const BalanceControl = ({game}) => {
    const [balances, setBalances] = useState(game.players.map((player) => player.balance));

    useEffect(() => {
        setBalances(game.players.map((player) => player.balance));
    }, [game]);

    const onChange = (event, index) => {
        setBalances([...balances.slice(0, index), event.target.value, ...balances.slice(index + 1, balances.length)])
    };

    const onSaveClick = () => {
        JEOPARDY_API.set_balance(balances.map(b => parseInt(b)));
    };

    return <div className={styles.balanceControl}>
        {game.players.map((player, index) => (
            <div key={index}>
                <div className={styles.name}>{player.name}</div>
                <Input type={"number"} onChange={(event) => onChange(event, index)} value={balances[index]}/>
            </div>
        ))}
        {game.players.length > 0 && <Button className={styles.save} onClick={onSaveClick}>Save</Button>}
    </div>
};

const Player = ({balance, name, onClick, selected}) => {
    return <div className={classNames(styles.button, selected && styles.selected, styles.player)} onClick={onClick}>
        <div>{balance}</div>
        <div>{name}</div>
    </div>
}

const stateContent = (game) => {
    const onSelectQuestion = (questionId) => {
        JEOPARDY_API.chooseQuestion(questionId);
    };
    const onSelectTheme = (themeId) => {
        JEOPARDY_API.remove_final_theme(themeId);
    };

    switch (game.state) {
        case "intro": return <TextContent>Intro</TextContent>
        case "themes_all": return <ThemesGrid game={game}/>
        case "round": return <TextContent>{getRoundName(game)}</TextContent>
        case "round_themes": return <ThemesList game={game}/>
        case "final_themes": return <ThemesList onSelect={onSelectTheme} game={game} active={true}/>
        case "questions": return <QuestionsGrid onSelect={onSelectQuestion} game={game}/>
        case "question_event":
            return <QuestionEvent question={game.question}/>
        case "question": case "answer": case "question_end": case "final_question":
            return <Question game={game}/>
        case "final_bets": return <FinalBets players={game.players}/>
        case "final_answer":
            return [
                <Question game={game} id={1}/>,
                <FinalAnswers players={game.players} id={2}/>
            ]
        case "final_player_answer": return <FinalAnswers players={game.players} answerer={game.answerer}/>
        case "final_player_bet": return <FinalBets players={game.players} answerer={game.answerer}/>
        case "game_end": return <TextContent>Game over</TextContent>
        default: return ""
    }
};

const control = (game, answerer, bet) => {
    const onNextClick = () => JEOPARDY_API.nextState(game.state);
    const onSkipClick = () => JEOPARDY_API.skip_question();
    const onAnswererClick = () => JEOPARDY_API.set_answerer_and_bet(answerer, bet);
    const onAnswerClick = (is_right) => JEOPARDY_API.answer(is_right);
    const onFinalAnswerClick = (is_right) => JEOPARDY_API.final_player_answer(is_right);

    const buttons = [];

    if (game.state === "question_event") {
        buttons.push(<Button onClick={onSkipClick}>Skip</Button>);
        if (answerer && bet > 0) {
            buttons.push(<Button onClick={onAnswererClick}>Next</Button>);
        }
    } else if (game.state === "answer") {
        buttons.push(<Button onClick={onSkipClick}>Skip</Button>);

        if (game.answerer) {
            buttons.push(<Button onClick={() => onAnswerClick(false)}>Wrong</Button>);
            buttons.push(<Button onClick={() => onAnswerClick(true)}>Right</Button>);
        }
    } else if (game.state ==="final_player_answer") {
        buttons.push(<Button onClick={() => onFinalAnswerClick(false)}>Wrong</Button>);
        buttons.push(<Button onClick={() => onFinalAnswerClick(true)}>Right</Button>);
    } else if (!["questions", "final_themes", "game_end"].includes(game.state)) {
        buttons.push(<Button onClick={onNextClick}>Next</Button>);
    }
    return buttons;
}

const JeopardyAdmin = () => {
    const game = useGame(JEOPARDY_API);
    const [connected, setConnected] = useAuth(JEOPARDY_API);
    const navigate = useNavigate();

    const [answerer, setAnswerer] = useState();
    const [bet, setBet] = useState();

    useEffect(() => {
        if (game) {
            setAnswerer(game.answerer);
            setBet(game.question ? game.question.value : 0);
        }
    }, [game]);

    const onSoundStop = () => JEOPARDY_API.intercom("sound_stop");
    const onReplay = () => {
        JEOPARDY_API.intercom("sound_stop");
        JEOPARDY_API.intercom("replay");
    }
    const onLogout = () => {
        JEOPARDY_API.logout();
        navigate("/admin");
    };
    const onPlayerSelect = (id) => {
        if (game.state === "question_event") setAnswerer(id);
    };

    if (!connected) return <AdminAuth api={JEOPARDY_API} setConnected={setConnected}/>;
    if (!game) return <Loading/>

    return <GameAdmin>
        <Header gameName={"Jeopardy"} token={game.token} stateName={getStatusName(game.state)}>
            <OvalButton onClick={onSoundStop}><FaVolumeMute /></OvalButton>
            <OvalButton onClick={onReplay}><MdReplayCircleFilled /></OvalButton>
            <ButtonLink to={"/admin"}>Home</ButtonLink>
            <ButtonLink to={"/jeopardy/view"}>View</ButtonLink>
            <Button onClick={onLogout}>Logout</Button>
        </Header>
        <Content rightPanel={<BalanceControl game={game}/>}>
            {stateContent(game)}
        </Content>
        <Footer>
            <FooterItem>
                <HorizontalList>
                    {game.players.map((player, index) => (
                        <TwoLineListItem
                            key={index} className={classNames(
                                game.state === "question_event" && styles.active,
                                styles.player, player.id === answerer && styles.selected)}
                            onClick={() => onPlayerSelect(player.id)}
                        >
                            <div>{player.balance}</div>
                            <div>{player.name}</div>
                        </TwoLineListItem>
                    ))}
                </HorizontalList>
                {game.state === "question_event" && <Input
                    className={classNames(styles.bet)} type={"number"}
                    onChange={e => setBet(parseInt(e.target.value))} value={bet}/>
                }
            </FooterItem>
            <FooterItem>
                {control(game, answerer, bet)}
            </FooterItem>
        </Footer>
    </GameAdmin>
}

export default JeopardyAdmin;
