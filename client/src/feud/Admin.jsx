import React from "react";
import {useNavigate} from "react-router-dom";

import {
    useGame,
    useAuth,
    Loading,
    Button,
    ButtonLink,
    VerticalList,
    ListItem,
    OvalButton,
    calcStateName
} from "common/Essentials"
import {Content, Footer, FooterItem, GameAdmin, Header, TextContent} from "common/Admin";
import {AdminAuth} from "common/Auth";

import styles from "feud/Admin.module.scss";
import {FinalQuestions, Question} from "feud/Question";
import {FaVolumeMute} from "react-icons/fa";
import classNames from "classnames";
import {FEUD_API} from "../index";


const getStateName = (state) => {
    return calcStateName(state);
}

const Players = ({game}) => {
    return <VerticalList className={styles.players}>
        {game.players.map(player =>
            <ListItem key={player.id} className={classNames(
                styles.player,
                player.id === game.answerer && styles.selected,
            )}>
                {player.name}
            </ListItem>
        )}
    </VerticalList>
};

const stateContent = (game) => {
    const onAnswerClick = (answerId) => FEUD_API.answer(true, answerId);

    switch (game.state) {
        case "round":
            return <TextContent>Round {game.round}</TextContent>;
        case "button":
        case "answers":
        case "answers_reveal":
        case "final_questions":
            return <Question
                game={game} showHiddenAnswers={true} className={styles.question} onSelect={onAnswerClick}
            />;
        case "final_questions_reveal":
            return <FinalQuestions game={game} className={styles.question} />
        default:
            return <TextContent>{getStateName(game.state)}</TextContent>;
    }
};

const control = (game) => {
    const onNextClick = () => FEUD_API.next_state(game.state);
    const onSetAnswererClick = (playerId) => FEUD_API.set_answerer(playerId);
    const onWrongAnswerClick = () => FEUD_API.answer(false, 0);

    const onRepeatClick = () => FEUD_API.intercom("repeat");

    const buttons = [];
    switch (game.state) {
        case "button":
            buttons.push(<Button key={2} onClick={() => onWrongAnswerClick()}>Wrong</Button>)
            buttons.push(game.players.map(player =>
                <Button key={100 + player.id} onClick={() => onSetAnswererClick(player.id)}>{player.name}</Button>,
            ));
            break;
        case "answers":
            buttons.push(<Button key={2} onClick={() => onWrongAnswerClick()}>Wrong</Button>)
            break;
        case "final":
            buttons.push(<Button key={1} onClick={() => onRepeatClick()}>Repeat</Button>)
            buttons.push(<Button key={5} onClick={onNextClick}>Next</Button>);
            break;
        case "final_questions":
            buttons.push(<Button key={1} onClick={() => onRepeatClick()}>Repeat</Button>)
            buttons.push(<Button key={2} onClick={() => onWrongAnswerClick()}>Wrong</Button>)
            break;
        case "end":
            break;
        default:
            buttons.push(<Button key={5} onClick={onNextClick}>Next</Button>);
    }
    return buttons;
};

const gameScore = (game) => {
    if (game.players.length < 2) return "";
    if (game.answerer && (game.state === 'final' || game.state === 'final_questions'
        || game.state === 'final_questions_reveal' || game.state === 'end')) {
        const answerer = game.answerer && game.players.find(t => t.id === game.answerer);
        return answerer.score + " | " + answerer.final_score;
    }
    return game.players[0].score + " : " + game.players[1].score;
}

const FeudAdmin = () => {
    const game = useGame(FEUD_API, (_) => {}, (_) => {});
    const [connected, setConnected] = useAuth(FEUD_API);
    const navigate = useNavigate();

    const onSoundStop = () => FEUD_API.intercom("sound_stop");
    const onLogout = () => {
        FEUD_API.logout();
        navigate("/admin");
    };

    if (!connected) return <AdminAuth api={FEUD_API} setConnected={setConnected}/>;
    if (!game) return <Loading/>;

    return <GameAdmin>
        <Header gameName={"Friends Feud"} token={game.token} stateName={getStateName(game.state)}>
            <OvalButton onClick={onSoundStop}><FaVolumeMute /></OvalButton>
            <ButtonLink to={"/admin"}>Home</ButtonLink>
            <ButtonLink to={"/feud/view"}>View</ButtonLink>
            <Button onClick={onLogout}>Logout</Button>
        </Header>
        <Content rightPanel={<Players game={game}/>}>
            {stateContent(game)}
        </Content>
        <Footer>
            <FooterItem className={styles.gameScore}>{gameScore(game)}</FooterItem>
            <FooterItem>{control(game)}</FooterItem>
        </Footer>
    </GameAdmin>
}

export default FeudAdmin;
