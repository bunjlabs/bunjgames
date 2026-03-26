import React from 'react';
import { useNavigate } from 'react-router-dom';
import { FaVolumeMute } from 'react-icons/fa';

import { Loading, Button, OvalButton, ButtonLink, VerticalList, ListItem } from 'components/UI';
import { useGame, useAuth, calcStateName } from 'components/hooks';
import { AdminAuth } from 'components/Auth';
import {
  GameAdmin, AdminHeader, AdminContent, TextContent,
  AdminFooter, FooterItem,
} from 'components/AdminLayout';
import { FinalQuestions, Question } from './Question';
import { FEUD_API } from './api';

const Players: React.FC<{ game: any }> = ({ game }) => (
  <VerticalList className="no-scrollbar" style={{ padding: 12, flex: '0 1 auto', minHeight: 0 } as any}>
    {game.players.map((p: any) => (
      <ListItem
        key={p.name}
        style={{
          backgroundColor: p.name === game.answerer?.name ? 'var(--bg-select)' : 'var(--bg-button)',
          color: p.name === game.answerer?.name ? 'var(--text-select)' : undefined,
          fontSize: 22, marginBottom: 5, padding: 5,
        } as any}
      >
        {p.name}
      </ListItem>
    ))}
  </VerticalList>
);

const stateContent = (game: any) => {
  switch (game.state) {
    case 'round': return <TextContent>Round {game.round}</TextContent>;
    case 'button': case 'answers': case 'answers_reveal': case 'final_questions':
      return <div style={{ padding: 12 }}>
        <Question game={game} showHiddenAnswers onSelect={(idx) => FEUD_API.answer(true, idx)} />
      </div>;
    case 'final_questions_reveal':
      return <div style={{ padding: 12 }}><FinalQuestions game={game} /></div>;
    default:
      return <TextContent>{calcStateName(game.state)}</TextContent>;
  }
};

const gameScore = (game: any) => {
  if (game.players.length < 2) return '';
  if (game.answerer && ['final', 'final_questions', 'final_questions_reveal', 'end'].includes(game.state)) {
    const a = game.players.find((t: any) => t.name === game.answerer.name);
    return a.score + ' | ' + a.finalScore;
  }
  return game.players[0].score + ' : ' + game.players[1].score;
};

const FeudAdmin: React.FC = () => {
  const game = useGame(FEUD_API, () => {}, () => {});
  const [connected, setConnected] = useAuth(FEUD_API);
  const navigate = useNavigate();

  const onLogout = () => { FEUD_API.logout(); navigate('/admin'); };

  if (!connected) return <AdminAuth api={FEUD_API} setConnected={setConnected} />;
  if (!game) return <Loading />;

  const onNext = () => FEUD_API.nextState(game.state);
  const onWrong = () => FEUD_API.answer(false, 0);
  const onRepeat = () => FEUD_API.intercom('repeat');

  const controls: React.ReactNode[] = [];
  switch (game.state) {
    case 'button':
      controls.push(<Button key="w" onClick={onWrong}>Wrong</Button>);
      game.players?.forEach((p: any) =>
        controls.push(<Button key={p.name} onClick={() => FEUD_API.setAnswerer(p.name)}>{p.name}</Button>)
      );
      break;
    case 'answers':
      controls.push(<Button key="w" onClick={onWrong}>Wrong</Button>);
      break;
    case 'final':
      controls.push(<Button key="rep" onClick={onRepeat}>Repeat</Button>);
      controls.push(<Button key="next" onClick={onNext}>Next</Button>);
      break;
    case 'final_questions':
      controls.push(<Button key="rep" onClick={onRepeat}>Repeat</Button>);
      controls.push(<Button key="w" onClick={onWrong}>Wrong</Button>);
      break;
    case 'end': break;
    default:
      controls.push(<Button key="next" onClick={onNext}>Next</Button>);
  }

  return (
    <GameAdmin>
      <AdminHeader gameName="Friends Feud" token={game.token} stateName={calcStateName(game.state)}>
        <OvalButton onClick={() => FEUD_API.intercom('sound_stop')}><FaVolumeMute /></OvalButton>
        <ButtonLink to="/admin">Home</ButtonLink>
        <ButtonLink to="/feud/view">View</ButtonLink>
        <Button onClick={onLogout}>Logout</Button>
      </AdminHeader>
      <AdminContent rightPanel={<Players game={game} />}>
        {stateContent(game)}
      </AdminContent>
      <AdminFooter>
        <FooterItem style={{ fontSize: 38, fontWeight: 'bold' }}>{gameScore(game)}</FooterItem>
        <FooterItem>{controls}</FooterItem>
      </AdminFooter>
    </GameAdmin>
  );
};

export default FeudAdmin;
