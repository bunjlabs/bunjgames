import React from 'react';

const questionBoxStyle = (correct: boolean): React.CSSProperties => ({
  height: 100, width: 100, marginRight: 10,
  backgroundColor: correct ? 'var(--bg-green)' : 'var(--bg-red)',
});

const emptyBoxStyle: React.CSSProperties = {
  height: 100, width: 100, marginRight: 10,
  backgroundColor: 'var(--bg-dark)',
};

const FinalQuestions: React.FC<{ game: any }> = ({ game }) => {
  const strongest = game.roundState.strongest;
  const weakest = game.roundState.weakest;
  if (!strongest || !weakest) return null;

  const strongestScore: boolean[] = strongest.finalScore ?? [];
  const weakestScore: boolean[] = weakest.finalScore ?? [];

  const maxLen = Math.max(strongestScore.length, weakestScore.length, 5);

  return (
    <div style={{ display: 'flex' }}>
      <div style={{ marginRight: 20 }}>
        <div style={{ height: 100, display: 'flex', alignItems: 'center', fontSize: 40, fontWeight: 'bold', minWidth: 0, maxWidth: 300, marginBottom: 10 }}>{strongest.name}</div>
        <div style={{ height: 100, display: 'flex', alignItems: 'center', fontSize: 40, fontWeight: 'bold', minWidth: 0, maxWidth: 300 }}>{weakest.name}</div>
      </div>
      <div>
        {[strongestScore, weakestScore].map((scores, qi) => (
          <div key={qi} style={{ display: 'flex', marginBottom: 10 }}>
            {Array.from({ length: maxLen }, (_, i) => (
              <div key={i} style={i < scores.length ? questionBoxStyle(scores[i]) : emptyBoxStyle} />
            ))}
          </div>
        ))}
      </div>
    </div>
  );
};

export default FinalQuestions;
