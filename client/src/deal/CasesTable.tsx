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

  const ordered = presents
    .map((p: any, i: number) => ({ p, i }))
    .filter(({ p }) => !p.given)
    .sort((a, b) => a.p.order - b.p.order);

  const count = ordered.length;
  const columns = Math.min(count, Math.max(2, Math.ceil(Math.sqrt(count))));
  const lastRowCount = count % columns;
  const lastRowOffset = lastRowCount > 0 && lastRowCount < columns
    ? Math.floor((columns - lastRowCount) / 2)
    : 0;
  const lastRowStart = count - lastRowCount;

  const cells = ordered.map(({ p, i }, idx) => {
    const isGiven = p.given;
    const isSelected = p.selected;
    const isOpened = p.opened;

    let label: React.ReactNode = idx + 1;
    if (showContent && !isGiven && !isSelected && !isOpened) {
      label = (
        <>
          <div>{idx + 1}</div>
          <div style={{ fontSize: 14, fontWeight: 'normal' }}>{p.content}</div>
        </>
      );
    }
    if (showContent && isSelected) {
      label = (
        <>
          <div>{idx + 1}</div>
          <div style={{ fontSize: 14, fontWeight: 'normal' }}>{p.content}</div>
        </>
      );
    }
    if (isOpened && !isGiven && !isSelected) {
      label = (
        <div style={{ color: 'var(--text-gray)' }}>{p.content}</div>
      );
    }
    if (isGiven) {
      label = null;
    }

    const clickable = onSelect && !isGiven && !isSelected && !isOpened;
    const isLastRow = lastRowOffset > 0 && idx >= lastRowStart;

    return (
      <div
        key={i}
        className={clickable ? 'clickable' : ''}
        style={{
          ...cellStyle,
          backgroundColor: isSelected ? 'var(--text-value)' : isGiven ? 'var(--bg-button)' : isOpened ? 'var(--bg-button)' : 'var(--bg-dark)',
          color: isSelected ? 'var(--text-select)' : 'var(--text)',
          borderColor: isSelected ? 'var(--bg-dark)' : isGiven ? 'transparent' : 'var(--bg-dark)',
          cursor: clickable ? 'pointer' : 'default',
          gridColumnStart: isLastRow ? idx - lastRowStart + 1 + lastRowOffset : undefined,
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
        gridTemplateColumns: `repeat(${columns}, minmax(0, 1fr))`,
        gridAutoRows: '1fr',
        gap: 10,
        height: '100%',
        width: '100%',
      }}
    >
      {cells}
    </div>
  );
};
