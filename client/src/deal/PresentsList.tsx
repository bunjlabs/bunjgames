import React from 'react';

const itemStyle: React.CSSProperties = {
  padding: '8px 12px',
  backgroundColor: 'var(--bg-dark)',
  color: 'var(--text)',
  fontSize: 20,
  whiteSpace: 'nowrap',
  overflow: 'hidden',
  textOverflow: 'ellipsis',
};

export const PresentsList: React.FC<{
  game: any;
  showDescription: boolean;
  onClick?: (index: number) => void;
}> = ({ game, showDescription, onClick }) => {
  const presents = game.presents as any[];

  const items = presents
    .map((p: any, i: number) => ({ p, i }))
    .filter(({ p }) => !p.given)
    .map(({ p, i }) => {
      const struck = p.opened;
      return (
        <div
          key={i}
          className={onClick ? 'clickable' : ''}
          style={{
            ...itemStyle,
            textDecoration: struck ? 'line-through' : 'none',
            cursor: onClick ? 'pointer' : 'default',
          }}
          onClick={() => onClick?.(i)}
        >
          <div>{p.content}</div>
          {showDescription && p.description && (
            <div style={{ fontSize: 14, fontWeight: 'normal', color: 'var(--text-gray)' }}>{p.description}</div>
          )}
        </div>
      );
    });

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 8, width: '100%' }}>
      {items}
    </div>
  );
};
