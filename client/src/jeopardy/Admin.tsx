import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { FaVolumeMute } from 'react-icons/fa';
import { MdReplayCircleFilled } from 'react-icons/md';

import { Loading, Button, OvalButton, ButtonLink, Input, HorizontalList, TwoLineListItem } from 'components/UI';
import { AudioPlayer, ImagePlayer, VideoPlayer } from 'components/Media';
import { useGame, useAuth } from 'components/hooks';
import { AdminAuth } from 'components/Auth';
import {
  GameAdmin, AdminHeader, AdminContent, BlockContent, TextContent,
  AdminFooter, FooterItem,
} from 'components/AdminLayout';
import { ThemesList, ThemesGrid, QuestionsGrid } from './Themes';
import { getStatusName, EventType, getRoundName } from './Common';
import { JEOPARDY_API } from './api';

const mediaStyle: React.CSSProperties = { whiteSpace: 'pre-wrap' };
const imgStyle: React.CSSProperties = { maxHeight: 300, objectFit: 'contain', width: '100%' };

const QuestionEvent: React.FC<{ question: any }> = ({ question }) => {
  const { type, theme, customTheme, value } = question;
  return (
    <BlockContent>
      <div>
        <div style={{ fontSize: 40 }}><EventType type={type} /></div>
        <div>{customTheme ? `Custom theme: ${customTheme}` : `Theme: ${theme}`}</div>
        <div>Value: {value}</div>
      </div>
    </BlockContent>
  );
};

const QuestionContent: React.FC<{ game: any }> = ({ game }) => {
  const question = game.state.question;
  const { value, customTheme, text, image, audio, video, answer, comment, answerText, answerImage, answerAudio, answerVideo } = question;
  return (
    <BlockContent>
      <div>
        <div>Value: {value}</div>
        {customTheme && <div>Custom theme: {customTheme}</div>}
      </div>
      <div style={mediaStyle}>
        {text && <p>{text}</p>}
        {image && <img style={imgStyle} src={`/media/${game.token}/${image}`} alt="" />}
        {audio && <AudioPlayer controls game={game} url={audio} />}
        {video && <VideoPlayer controls game={game} url={video} />}
      </div>
      <div style={mediaStyle}>
        <div>{answer}</div>
        {comment && <div>Comment: {comment}</div>}
        {answerText && <p>{answerText}</p>}
        {answerImage && <ImagePlayer game={game} url={answerImage} />}
        {answerAudio && <AudioPlayer controls game={game} url={answerAudio} />}
        {answerVideo && <VideoPlayer controls game={game} url={answerVideo} />}
      </div>
    </BlockContent>
  );
};

const FinalBets: React.FC<{ players: any[]; answerer?: string }> = ({ players, answerer }) => (
  <div style={{ padding: 8 }}>
    {players.map((p, i) => (
      <PlayerItem key={i} balance={p.finalBet} name={p.name} selected={answerer === p.name}
        onClick={() => JEOPARDY_API.intercom('do_bet:' + p.name)} />
    ))}
  </div>
);

const FinalAnswers: React.FC<{ players: any[]; answerer?: string }> = ({ players, answerer }) => (
  <div style={{ padding: 8 }}>
    {players.map((p, i) => (
      <PlayerItem key={i} balance={p.finalAnswer || '\u2E3B'} name={p.name} selected={answerer === p.name}
        onClick={() => JEOPARDY_API.intercom('do_answer:' + p.name)} />
    ))}
  </div>
);

const PlayerItem: React.FC<{
  balance: any; name: string; onClick: () => void; selected?: boolean;
}> = ({ balance, name, onClick, selected }) => (
  <div
    className="btn"
    style={{
      minWidth: 150, padding: 8, marginBottom: 8,
      backgroundColor: selected ? 'var(--bg-select)' : 'var(--bg-button)',
      color: selected ? 'var(--text-select)' : 'var(--text)',
    }}
    onClick={onClick}
  >
    <div>{balance}</div>
    <div>{name}</div>
  </div>
);

const RoundSelector: React.FC<{ game: any }> = ({ game }) => {
  const currentRound = game.round?.number ?? 1;
  const roundCount = game.roundCount ?? 1;

  const options = Array.from({ length: roundCount }, (_, i) => {
    const num = i + 1;
    const label = num === roundCount && roundCount > 1 ? 'FINAL' : `ROUND ${num}`;
    return { value: num, label };
  });

  const onChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const selected = parseInt(e.target.value);
    if (selected !== currentRound) {
      JEOPARDY_API.setRound(selected);
    }
  };

  return (
    <div style={{ borderTop: '8px solid var(--bg-dark)', padding: 10 }}>
      <div style={{ fontSize: 18, marginBottom: 8, color: 'var(--text-white)' }}>Change Round</div>
      <select
        className="input"
        value={currentRound}
        onChange={onChange}
        style={{ width: '100%', cursor: 'pointer' }}
      >
        {options.map((o) => (
          <option key={o.value} value={o.value}>{o.label}</option>
        ))}
      </select>
    </div>
  );
};

const BalanceControl: React.FC<{ game: any }> = ({ game }) => {
  const [balances, setBalances] = useState(game.players.map((p: any) => p.balance));

  useEffect(() => {
    setBalances(game.players.map((p: any) => p.balance));
  }, [game]);

  const onChange = (val: string, idx: number) => {
    setBalances([...balances.slice(0, idx), val, ...balances.slice(idx + 1)]);
  };

  return (
    <div style={{ padding: 10 }}>
      {game.players.map((p: any, i: number) => (
        <div key={i} style={{ marginBottom: 16 }}>
          <div style={{ fontSize: 22 }}>{p.name}</div>
          <Input type="number" onChange={(e) => onChange(e.target.value, i)} value={balances[i]} />
        </div>
      ))}
      {game.players.length > 0 && (
        <Button style={{ fontSize: 22 }} onClick={() => JEOPARDY_API.setBalance(balances.map((b: any) => parseInt(b)))}>
          Save
        </Button>
      )}
    </div>
  );
};

const stateContent = (game: any) => {
  switch (game.state.value) {
    case 'intro': return <TextContent>Intro</TextContent>;
    case 'themes_all': return <ThemesGrid game={game} />;
    case 'round': return <TextContent>{getRoundName(game)}</TextContent>;
    case 'round_themes': return <ThemesList game={game} />;
    case 'final_themes': return <ThemesList onSelect={(name) => JEOPARDY_API.removeFinalTheme(name)} game={game} active />;
    case 'questions': return <QuestionsGrid onSelect={(q) => JEOPARDY_API.chooseQuestion(q)} game={game} />;
    case 'question_event': return <QuestionEvent question={game.state.question} />;
    case 'question': case 'answer': case 'question_end': case 'final_question':
      return <QuestionContent game={game} />;
    case 'final_bets': return <FinalBets players={game.players} />;
    case 'final_answer':
      return <><QuestionContent game={game} /><FinalAnswers players={game.players} /></>;
    case 'final_player_answer': return <FinalAnswers players={game.players} answerer={game.state.answerer?.name} />;
    case 'final_player_bet': return <FinalBets players={game.players} answerer={game.state.answerer?.name} />;
    case 'game_end': return <TextContent>Game over</TextContent>;
    default: return null;
  }
};

const JeopardyAdmin: React.FC = () => {
  const game = useGame(JEOPARDY_API);
  const [connected, setConnected] = useAuth(JEOPARDY_API);
  const navigate = useNavigate();
  const [answerer, setAnswerer] = useState<string>();
  const [bet, setBet] = useState<number>(0);

  useEffect(() => {
    if (game) {
      setAnswerer(game.state.answerer?.name);
      setBet(game.state.question ? game.state.question.value : 0);
    }
  }, [game]);

  const onLogout = () => { JEOPARDY_API.logout(); navigate('/admin'); };

  if (!connected) return <AdminAuth api={JEOPARDY_API} setConnected={setConnected} />;
  if (!game) return <Loading />;

  const onNext = () => JEOPARDY_API.nextState(game.state.value);
  const onSkip = () => JEOPARDY_API.skipQuestion();
  const onSetAnswerer = () => answerer && JEOPARDY_API.setAnswererAndBet(answerer, bet);
  const onAnswer = (ok: boolean) => JEOPARDY_API.answer(ok);
  const onFinalAnswer = (ok: boolean) => JEOPARDY_API.finalPlayerAnswer(ok);
  const onPlayerSelect = (name: string) => { if (game.state.value === 'question_event') setAnswerer(name); };

  const controls = [];
  if (game.state.value === 'question_event') {
    controls.push(<Button key="skip" onClick={onSkip}>Skip</Button>);
    if (answerer && bet > 0) controls.push(<Button key="next" onClick={onSetAnswerer}>Next</Button>);
  } else if (game.state.value === 'answer') {
    controls.push(<Button key="skip" onClick={onSkip}>Skip</Button>);
    if (game.state.answerer) {
      controls.push(<Button key="w" onClick={() => onAnswer(false)}>Wrong</Button>);
      controls.push(<Button key="r" onClick={() => onAnswer(true)}>Right</Button>);
    }
  } else if (game.state.value === 'final_player_answer') {
    controls.push(<Button key="w" onClick={() => onFinalAnswer(false)}>Wrong</Button>);
    controls.push(<Button key="r" onClick={() => onFinalAnswer(true)}>Right</Button>);
  } else if (!['questions', 'final_themes', 'game_end'].includes(game.state.value)) {
    controls.push(<Button key="next" onClick={onNext}>Next</Button>);
  }

  const playerStyle = (p: any): React.CSSProperties => ({
    backgroundColor: p.name === answerer ? 'var(--bg-select)' : 'var(--bg-button)',
    color: p.name === answerer ? 'var(--text-select)' : undefined,
    minWidth: 150, padding: 8,
    cursor: game.state.value === 'question_event' ? 'pointer' : undefined,
  });

  return (
    <GameAdmin>
      <AdminHeader gameName="Jeopardy" token={game.token} stateName={getStatusName(game.state.value)}>
        <OvalButton onClick={() => JEOPARDY_API.intercom('sound_stop')}><FaVolumeMute /></OvalButton>
        <OvalButton onClick={() => { JEOPARDY_API.intercom('sound_stop'); JEOPARDY_API.intercom('replay'); }}>
          <MdReplayCircleFilled />
        </OvalButton>
        <ButtonLink to="/admin">Home</ButtonLink>
        <ButtonLink to="/jeopardy/view">View</ButtonLink>
        <Button onClick={onLogout}>Logout</Button>
      </AdminHeader>
      <AdminContent rightPanel={<><BalanceControl game={game} /><RoundSelector game={game} /></>}>
        {stateContent(game)}
      </AdminContent>
      <AdminFooter>
        <FooterItem>
          <HorizontalList>
            {game.players?.map((p: any, i: number) => (
              <TwoLineListItem key={i} className={playerStyle(p).backgroundColor === 'var(--bg-select)' ? 'selected' : ''}
                onClick={() => onPlayerSelect(p.name)}
              >
                <div>{p.balance}</div>
                <div>{p.name}</div>
              </TwoLineListItem>
            ))}
          </HorizontalList>
          {game.state.value === 'question_event' && (
            <Input className="" type="number" style={{ marginLeft: 32, maxWidth: 200 } as any}
              onChange={(e) => setBet(parseInt(e.target.value))} value={bet} />
          )}
        </FooterItem>
        <FooterItem>{controls}</FooterItem>
      </AdminFooter>
    </GameAdmin>
  );
};

export default JeopardyAdmin;
