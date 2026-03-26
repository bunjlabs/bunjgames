import React from 'react';
import { FaTimes } from 'react-icons/fa';

const cellBase: React.CSSProperties = {
  fontSize: 38, textAlign: 'center', display: 'flex', flexDirection: 'column',
  justifyContent: 'center', alignItems: 'center',
};

export const Question: React.FC<{
  game: any; showHiddenAnswers: boolean; className?: string; onSelect?: (index: number) => void;
}> = ({ game, showHiddenAnswers, className, onSelect }) => {
  const answerer = game.answerer && game.players.find((t: any) => t.name === game.answerer?.name);
  const answers: React.ReactNode[] = [];
  let strikes = 0;

  game.question.answers.forEach((a: any, i: number) => {
    const opened = game.state !== 'final_questions' ? a.isOpened : showHiddenAnswers && a.isFinalAnswered;

    answers.push(
      <div
        key={`a_${i}`}
        style={{ ...cellBase, backgroundColor: 'var(--bg-dark)', color: opened ? 'var(--text)' : 'var(--text-gray)', cursor: !opened && onSelect ? 'pointer' : undefined }}
        className={!opened && onSelect ? 'clickable' : ''}
        onClick={() => !opened && onSelect?.(i)}
      >
        {(opened || showHiddenAnswers) && a.text}
      </div>,
    );

    answers.push(
      <div key={`v_${i}`} style={{ ...cellBase, backgroundColor: 'var(--bg-dark)', color: opened ? 'var(--text)' : 'var(--text-gray)', fontWeight: 'bold' }}>
        {(opened || showHiddenAnswers) && a.value}
      </div>,
    );

    if (answerer && strikes < 3) {
      strikes++;
      answers.push(
        <div key={`s_${i}`} style={{
          ...cellBase, fontWeight: 'bold', fontSize: 56,
          backgroundColor: answerer.strikes >= strikes ? 'var(--bg-red)' : 'var(--bg-dark)',
          color: answerer.strikes >= strikes ? 'var(--text)' : 'var(--text-gray)',
        }}>
          <FaTimes />
        </div>,
      );
    } else {
      answers.push(<div key={`s_${i}`} />);
    }
  });

  return (
    <div className={className} style={{ display: 'grid', gridTemplateColumns: '3fr 1fr 0.7fr', gridAutoRows: '1fr', gap: 10, height: '100%', width: '100%' }}>
      <div style={{ ...cellBase, gridColumn: '1 / 4', marginBottom: 10, color: 'var(--text)' }}>{game.question.text}</div>
      {answers}
    </div>
  );
};

export const FinalQuestions: React.FC<{ game: any; className?: string }> = ({ game, className }) => {
  const answers: React.ReactNode[] = [];

  game.finalQuestions.forEach((q: any, i: number) => {
    answers.push(
      <div key={`q_${i}`} style={{ ...cellBase, backgroundColor: 'var(--bg-dark)', color: 'var(--text)' }}>
        {!q.isProcessed && q.text}
      </div>,
    );
    const a = q.answers.length ? q.answers[0] : null;
    answers.push(
      <div key={`a_${i}`} style={{ ...cellBase, backgroundColor: 'var(--bg-dark)', color: 'var(--text)' }}>
        {!q.isProcessed && (a ? a.text : '-')}
      </div>,
    );
    answers.push(
      <div key={`v_${i}`} style={{ ...cellBase, backgroundColor: 'var(--bg-dark)', color: 'var(--text)', fontWeight: 'bold' }}>
        {!q.isProcessed && (a ? a.value : '0')}
      </div>,
    );
  });

  return (
    <div className={className} style={{ display: 'grid', gridTemplateColumns: '3fr 2fr 1fr', gridAutoRows: '1fr', gap: 10, height: '100%', width: '100%' }}>
      {answers}
    </div>
  );
};
