import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { FaCheckSquare, FaMinus, FaPlus, FaSquare, FaVolumeMute } from 'react-icons/fa';

import { Loading } from 'components/UI';
import { Button, OvalButton, ButtonLink } from 'components/UI';
import { ImagePlayer, AudioPlayer, VideoPlayer } from 'components/Media';
import { useGame, useAuth } from 'components/hooks';
import { AdminAuth } from 'components/Auth';
import {
  GameAdmin, AdminHeader, AdminContent, BlockContent, TextContent,
  AdminFooter, FooterItem,
} from 'components/AdminLayout';
import { WHIRLIGIG_API } from './api';

const STATUS_NAMES: Record<string, string> = {
  start: 'Start', intro: 'Intro', questions: 'Questions',
  question_whirligig: 'Selecting question', question_start: 'Asking',
  question_discussion: 'Discussion', answer: 'Answer',
  right_answer: 'Right answer', question_end: 'Question end', end: 'Game over',
};

const getStatusName = (s: string) => STATUS_NAMES[s] ?? '';

const ItemQuestion: React.FC<{ question: any; single: boolean }> = ({ question, single }) => {
  const checkbox = question.isProcessed ? <FaCheckSquare /> : <FaSquare />;
  return (
    <div style={{ padding: 10 }}>
      <div>
        {!single && <span style={{ float: 'right', margin: 10 }}>{checkbox}</span>}
        Question: {question.description}
      </div>
      <div>Answer: {question.answer.description}</div>
    </div>
  );
};

const Item: React.FC<{ item: any }> = ({ item }) => {
  const [open, setOpen] = useState(false);
  const checkbox = item.is_processed ? <FaCheckSquare /> : <FaSquare />;
  return (
    <div style={{ color: 'var(--text)', fontSize: 18, marginBottom: 10 }}>
      <div
        style={{ display: 'flex', cursor: 'pointer' }}
        onClick={() => setOpen(!open)}
      >
        <div style={{ flexGrow: 1 }}>{item.description || item.name}</div>
        <div>{checkbox}</div>
      </div>
      {open && item.questions.map((q: any, k: number) => (
        <ItemQuestion key={k} question={q} single={item.questions.length <= 1} />
      ))}
    </div>
  );
};

const Items: React.FC<{ items: any[] }> = ({ items }) => (
  <div className="no-scrollbar" style={{ padding: 10, flex: '0 1 auto', minHeight: 0 }}>
    {items.map((item, i) => <Item key={i} item={item} />)}
  </div>
);

const Timer: React.FC<{ game: any }> = ({ game }) => {
  const [time, setTime] = useState(Math.max(Math.floor(game.timer.time / 1000), 0));

  useEffect(() => {
    const fromTime = Date.now();
    const timer = setInterval(() => setTime(WHIRLIGIG_API.getTime(fromTime)), 100);
    return () => clearInterval(timer);
  }, [game]);

  return (
    <div>
      <div style={{ textAlign: 'center', fontSize: 30 }}>{time}</div>
      <Button
        style={{ fontSize: 18, height: 28, width: 100 }}
        onClick={() => WHIRLIGIG_API.timer(!game.timer.paused)}
      >
        {game.timer.paused ? 'Resume' : 'Pause'}
      </Button>
    </div>
  );
};

const ScoreControl: React.FC<{ game: any }> = ({ game }) => {
  const update = (c: number, v: number) => WHIRLIGIG_API.score(c, v);
  const { connoisseurs, viewers } = game.score;

  const controlStyle: React.CSSProperties = { textAlign: 'center', marginBottom: 10 };
  const rowStyle: React.CSSProperties = { display: 'flex', justifyContent: 'center', alignItems: 'center', margin: '0 10px', gap: 10 };
  const smallBtn: React.CSSProperties = { height: 30, width: 30, padding: 0 };

  return (
    <div style={{ color: 'var(--text)', borderTop: '10px solid var(--bg-dark)', padding: 10 }}>
      <div style={controlStyle}>
        <div>Connoisseurs</div>
        <div style={rowStyle}>
          <Button style={smallBtn} onClick={() => update(connoisseurs - 1, viewers)}><FaMinus /></Button>
          {connoisseurs}
          <Button style={smallBtn} onClick={() => update(connoisseurs + 1, viewers)}><FaPlus /></Button>
        </div>
      </div>
      <div style={controlStyle}>
        <div>Viewers</div>
        <div style={rowStyle}>
          <Button style={smallBtn} onClick={() => update(connoisseurs, viewers - 1)}><FaMinus /></Button>
          {viewers}
          <Button style={smallBtn} onClick={() => update(connoisseurs, viewers + 1)}><FaPlus /></Button>
        </div>
      </div>
    </div>
  );
};

const StateContent: React.FC<{ game: any }> = ({ game }) => {
  const item = game.state.item;
  const question = game.state.question;
  const mediaStyle: React.CSSProperties = { whiteSpace: 'pre-wrap' };
  const mediaSize: React.CSSProperties = { maxWidth: 600, maxHeight: 400 };

  const QuestionInfo = () => (
    <>
      <div>{getStatusName(game.state.value)}</div>
      <div>
        <div>Name: {item.name}</div>
        {item.description && <div>Description: {item.description}</div>}
        <div>Type: {item.type}</div>
      </div>
      {question.author && <div>Author: {question.author}</div>}
    </>
  );

  if (game.state.value === 'questions') {
    return (
      <BlockContent>
        {game.items.map((it: any, i: number) => (
          <div key={i}>
            {it.questions.length === 1 && Boolean(it.questions[0].author)
              ? (i === 12 ? '13 - ' : '') + it.questions[0].author
              : it.name}
          </div>
        ))}
        <div />
      </BlockContent>
    );
  }

  if (!item || !question) return <TextContent>{getStatusName(game.state.value)}</TextContent>;

  if (game.state.value === 'question_whirligig') {
    return <BlockContent><QuestionInfo /></BlockContent>;
  }

  return (
    <BlockContent>
      <div><QuestionInfo /></div>
      <div style={mediaStyle}><div>{question.description}</div></div>
      {(question.text || question.image || question.audio || question.video) && (
        <div style={mediaStyle}>
          {question.text && <p>{question.text}</p>}
          {question.image && <ImagePlayer game={game} url={question.image} style={mediaSize} />}
          {question.audio && <AudioPlayer controls game={game} url={question.audio} />}
          {question.video && <VideoPlayer controls game={game} url={question.video} style={mediaSize} />}
        </div>
      )}
      <div style={mediaStyle}>
        <div>{question.answer.description}</div>
        {question.answer.text && <p>{question.answer.text}</p>}
        {question.answer.image && <ImagePlayer game={game} url={question.answer.image} style={mediaSize} />}
        {question.answer.audio && <AudioPlayer controls game={game} url={question.answer.audio} />}
        {question.answer.video && <VideoPlayer controls game={game} url={question.answer.video} style={mediaSize} />}
      </div>
    </BlockContent>
  );
};

const Controls: React.FC<{ game: any }> = ({ game }) => {
  const onGong = () => WHIRLIGIG_API.intercom('gong');
  const onAnswer = (ok: boolean) => WHIRLIGIG_API.answerCorrect(ok);
  const onNext = () => WHIRLIGIG_API.nextState(game.state.value);
  const onExtra = () => WHIRLIGIG_API.extraTime();

  return (
    <>
      {game.state.value === 'question_discussion' && <Timer game={game} />}
      {game.state.value === 'answer' && <Button onClick={onExtra}>+time</Button>}
      <Button onClick={onGong} style={{ width: 64, height: 64, borderRadius: '50%', fontSize: 18 }}>Gong</Button>
      {game.state.value === 'right_answer' ? (
        <>
          <Button onClick={() => onAnswer(false)}>Wrong</Button>
          <Button onClick={() => onAnswer(true)}>Right</Button>
        </>
      ) : game.state.value !== 'end' ? (
        <Button onClick={onNext}>Next</Button>
      ) : null}
    </>
  );
};

const WhirligigAdmin: React.FC = () => {
  const game = useGame(WHIRLIGIG_API);
  const [connected, setConnected] = useAuth(WHIRLIGIG_API);
  const navigate = useNavigate();

  const onSoundStop = () => WHIRLIGIG_API.intercom('sound_stop');
  const onLogout = () => { WHIRLIGIG_API.logout(); navigate('/admin'); };

  if (!connected) return <AdminAuth api={WHIRLIGIG_API} setConnected={setConnected} />;
  if (!game) return <Loading />;

  return (
    <GameAdmin>
      <AdminHeader gameName="Whirligig" token={game.token} stateName={getStatusName(game.state.value)}>
        <OvalButton onClick={onSoundStop}><FaVolumeMute /></OvalButton>
        <ButtonLink to="/admin">Home</ButtonLink>
        <ButtonLink to="/whirligig/view">View</ButtonLink>
        <Button onClick={onLogout}>Logout</Button>
      </AdminHeader>
      <AdminContent rightPanel={<><Items items={game.items || []} /><ScoreControl game={game} /></>}>
        <StateContent game={game} />
      </AdminContent>
      <AdminFooter>
        <FooterItem style={{ fontSize: 38, fontWeight: 'bold' }}>{game.score.connoisseurs} : {game.score.viewers}</FooterItem>
        <FooterItem><Controls game={game} /></FooterItem>
      </AdminFooter>
    </GameAdmin>
  );
};

export default WhirligigAdmin;
