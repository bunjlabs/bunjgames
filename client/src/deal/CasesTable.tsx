import React from 'react';

const cellStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  alignItems: 'center',
  justifyContent: 'center',
  border: '4px solid var(--bg-dark)',
  borderRadius: 8,
  backgroundColor: 'var(--bg-dark)',
  color: 'var(--text)',
  fontSize: 28,
  fontWeight: 'bold',
  minHeight: 70,
  cursor: 'default',
  userSelect: 'none',
};

export const CasesTable: React.FC<{
  game: any;
  showContent: boolean;
  onSelect?: (index: number) => void;
}> = ({ game, showContent, onSelect }) => {
  const presents = game.presents as any[];

  const cells = presents.map((p: any, i: number) => {
    const isGiven = p.given;
    const isSelected = p.selected;
    const isOpened = p.opened;

    let label: React.ReactNode = i + 1;
    if (showContent && !isGiven && !isSelected && !isOpened) {
      label = (
        <>
          <div>{i + 1}</div>
          <div style={{ fontSize: 14, fontWeight: 'normal' }}>{p.content}</div>
        </>
      );
    }
    if (isOpened && !isGiven) {
      label = (
        <div style={{ color: 'var(--text-gray)' }}>{p.content}</div>
      );
    }
    if (isSelected) {
      label = null;
    }
    if (isGiven) {
      label = null;
    }

    const blank = isGiven || isSelected;
    const clickable = onSelect && !isGiven && !isSelected && !isOpened;

    return (
      <div
        key={i}
        className={clickable ? 'clickable' : ''}
        style={{
          ...cellStyle,
          backgroundColor: blank ? 'transparent' : isOpened ? 'var(--bg-button)' : 'var(--bg-dark)',
          borderColor: blank ? 'transparent' : 'var(--bg-dark)',
          cursor: clickable ? 'pointer' : 'default',
        }}
        onClick={() => clickable && onSelect?.(i)}
      >
        {label}
      </div>
    );
  });

  return (
    <div
      style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fill, minmax(90px, 1fr))',
        gap: 10,
        width: '100%',
      }}
    >
      {cells}
    </div>
  );
};
