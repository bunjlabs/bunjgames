import React, {useEffect, useState} from 'react';
import { useNavigate } from 'react-router-dom';
import { FaVolumeMute } from 'react-icons/fa';

import { Loading, Button, OvalButton, ButtonLink, VerticalList, ListItem, TwoLineListItem } from 'components/UI';
import { useGame, useAuth, useTimer } from 'components/hooks';
import { AdminAuth } from 'components/Auth';
import {
  GameAdmin, AdminHeader, AdminContent, BlockContent, TextContent,
  AdminFooter, FooterItem,
} from 'components/AdminLayout';
import FinalQuestions from './FinalQuestions';
import { WEAKEST_API } from './api';

const STATUS_NAMES: Record<string, string> = {
  waiting_for_players: 'Waiting for players', intro: 'Intro', round: 'Round',
  questions: 'Questions', weakest_choose: 'Weakest choose',
  weakest_reveal: 'Weakest reveal', final: 'Final',
  final_questions: 'Final questions', end: 'Game over',
};

const getStatusName = (s: string) => STATUS_NAMES[s] ?? '';

const Timer: React.FC<{ game: any }> = ({ game }) => {
  const [time, setTime] = useState(Math.max(Math.floor(game.roundState.time / 1000), 0));

  useEffect(() => {
    const fromTime = Date.now();
    const timer = setInterval(() => setTime(WEAKEST_API.getTime(fromTime)), 100);
    return () => clearInterval(timer);
  }, [game]);

  return <TextContent>{Math.floor((time % 3600) / 60000)}:{Math.floor(time % 60) / 1000}</TextContent>;
};

const Question: React.FC<{ game: any }> = ({ game }) => (
  <BlockContent>
    {game.state.value === 'final_questions' && <FinalQuestions game={game} />}
    {game.state.value === 'questions' && <Timer game={game}/>}
    <div>{game.roundState.question?.question}</div>
    <div>{game.roundState.question?.answer}</div>
  </BlockContent>
);

const WeakestContent: React.FC<{ game: any }> = ({ game }) => {
  const weakest = game.roundState.weakest;
  const strongest = game.roundState.strongest;
  return (
    <BlockContent>
      {game.state.value === 'weakest_reveal' && <TextContent>Weakest reveal</TextContent>}
      {weakest && <div>Weakest: {weakest.name}, answers: {weakest.rightAnswers}, income: {weakest.bankIncome}</div>}
      {strongest && <div>Strongest: {strongest.name}, answers: {strongest.rightAnswers}, income: {strongest.bankIncome}</div>}
      <VerticalList style={{ padding: 12 } as any}>
        {game.players.filter((p: any) => p.active).map((p: any) => (
          <TwoLineListItem key={p.name} style={{ backgroundColor: 'var(--bg-button)', padding: 5 }}>
            <div>{p.vote || '\u2E3B'}</div>
            <div>{p.name}</div>
          </TwoLineListItem>
        ))}
      </VerticalList>
    </BlockContent>
  );
};

const Players: React.FC<{ game: any }> = ({ game }) => (
  <VerticalList className="no-scrollbar" style={{ padding: 12, flex: '0 1 auto', minHeight: 0 } as any}>
    {game.players.map((p: any) => (
      <ListItem
        key={p.name}
        style={{
          backgroundColor: game.roundState.answerer?.name === p.name ? 'var(--bg-select)' : 'var(--bg-button)',
          color: game.roundState.answerer?.name === p.name ? 'var(--text-select)' : !p.active ? 'var(--text-gray)' : undefined,
          fontSize: 22, marginBottom: 5, padding: 5,
        } as any}
      >
        {p.name}
      </ListItem>
    ))}
  </VerticalList>
);

const ScoreList: React.FC<{ game: any }> = ({ game }) => {
  const scores = [40, 30, 20, 15, 10, 5, 2, 1];
  return (
    <VerticalList style={{ borderTop: '10px solid var(--bg-dark)', padding: 12, alignItems: 'center' } as any}>
      {scores.map((s) => (
        <ListItem
          key={s}
          style={{
            backgroundColor: game.state.value === 'questions' && game.roundState.score === s ? 'var(--bg-select)' : 'var(--bg-button)',
            color: game.state.value === 'questions' && game.roundState.score === s ? 'var(--text-select)' : undefined,
            fontSize: 18, marginBottom: 5, padding: 5, minWidth: 150,
          } as any}
        >
          {s * game.scoreMultiplier}
        </ListItem>
      ))}
    </VerticalList>
  );
};

const stateContent = (game: any) => {
  switch (game.state.value) {
    case 'intro': return <TextContent>Intro</TextContent>;
    case 'round': return <TextContent>Round {game.roundState.number}</TextContent>;
    case 'questions': case 'final_questions': return <Question game={game} />;
    case 'weakest_choose': case 'weakest_reveal': return <WeakestContent game={game} />;
    case 'final': return <TextContent>Choose player to start</TextContent>;
    case 'end': return <TextContent>Game over</TextContent>;
    default: return null;
  }
};

const WeakestAdmin: React.FC = () => {
  const game = useGame(WEAKEST_API, () => {}, () => {});
  const [connected, setConnected] = useAuth(WEAKEST_API);
  const navigate = useNavigate();

  const onLogout = () => { WEAKEST_API.logout(); navigate('/admin'); };

  if (!connected) return <AdminAuth api={WEAKEST_API} setConnected={setConnected} />;
  if (!game) return <Loading />;

  const state = game.state.value;
  const onNext = () => WEAKEST_API.nextState(state);
  const onAnswer = (ok: boolean) => WEAKEST_API.answerCorrect(ok);
  const onBank = () => WEAKEST_API.saveBank();
  const onGong = () => WEAKEST_API.intercom('gong');

  const controls = [<Button key="gong" onClick={onGong}>Gong</Button>];
  if (['questions', 'final_questions'].includes(state)) {
    controls.push(
      <Button key="bank" onClick={onBank}>Bank</Button>,
      <Button key="w" onClick={() => onAnswer(false)}>Wrong</Button>,
      <Button key="r" onClick={() => onAnswer(true)}>Right</Button>,
    );
  } else if (state === 'final') {
    game.players?.filter((p: any) => p.active).forEach((p: any) =>
      controls.push(<Button key={p.name} onClick={() => WEAKEST_API.selectFinalAnswerer(p.name)}>{p.name}</Button>)
    );
  } else if (!['end'].includes(state)) {
    controls.push(<Button key="next" onClick={onNext}>Next</Button>);
  }

  return (
    <GameAdmin>
      <AdminHeader gameName="The Weakest" token={game.token} stateName={getStatusName(state)}>
        <OvalButton onClick={() => WEAKEST_API.intercom('sound_stop')}><FaVolumeMute /></OvalButton>
        <ButtonLink to="/admin">Home</ButtonLink>
        <ButtonLink to="/weakest/view">View</ButtonLink>
        <Button onClick={onLogout}>Logout</Button>
      </AdminHeader>
      <AdminContent rightPanel={<><Players game={game} /><ScoreList game={game} /></>}>
        {stateContent(game)}
      </AdminContent>
      <AdminFooter>
        <FooterItem style={{ fontSize: 38, fontWeight: 'bold' }}>{game.state.score} ; {game.roundState.bank}</FooterItem>
        <FooterItem>{controls}</FooterItem>
      </AdminFooter>
    </GameAdmin>
  );
};

export default WeakestAdmin;
