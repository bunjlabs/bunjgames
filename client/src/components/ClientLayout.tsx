import React from 'react';
import { FaTimesCircle } from 'react-icons/fa';
import { Toast } from './UI';

export const ClientExitButton: React.FC<{ onClick: () => void }> = ({ onClick }) => (
  <button
    style={{ padding: 10, fontSize: 30, color: 'var(--text)' }}
    onClick={(e) => {
      if (window.confirm('Are you sure want to exit?')) onClick();
      else e.preventDefault();
    }}
  >
    <FaTimesCircle />
  </button>
);

export const ClientHeader: React.FC<React.PropsWithChildren> = ({ children }) => (
  <div style={{ position: 'absolute', top: 10 }}>{children}</div>
);

export const ClientTextContent: React.FC<React.PropsWithChildren> = ({ children }) => (
  <div style={{ fontSize: 40, fontWeight: 'bold' }}>{children}</div>
);

export const ClientFormContent: React.FC<React.PropsWithChildren> = ({ children }) => (
  <div>{children}</div>
);

export const BigButtonContent: React.FC<React.PropsWithChildren<{
  active?: boolean;
  onClick?: () => void;
}>> = ({ active, onClick, children }) => (
  <div
    style={{
      width: '90vmin', height: '90vmin', borderRadius: '50%',
      border: '4px solid var(--bg-dark)',
      backgroundColor: active ? '#c63939' : 'var(--bg-button)',
      boxShadow: '0 8px 0 var(--bg-dark)',
      display: 'flex', justifyContent: 'center', alignItems: 'center', fontSize: 40,
    }}
    onClick={onClick}
    onTouchStart={onClick}
  >
    {children}
  </div>
);

export const ClientContent: React.FC<React.PropsWithChildren> = ({ children }) => (
  <div style={{ display: 'flex', flexGrow: 1, justifyContent: 'center', alignItems: 'center' }}>{children}</div>
);

const useOrientation = () => {
  const [isLandscape, setLandscape] = React.useState(
    window.matchMedia('(orientation: landscape)').matches,
  );
  React.useEffect(() => {
    const mq = window.matchMedia('(orientation: landscape)');
    const handler = (e: MediaQueryListEvent) => setLandscape(e.matches);
    mq.addEventListener('change', handler);
    return () => mq.removeEventListener('change', handler);
  }, []);
  return isLandscape;
};

export const GameClient: React.FC<React.PropsWithChildren> = ({ children }) => {
  const isLandscape = useOrientation();
  return (
    <div style={{
      height: '100%', width: '100%', backgroundColor: 'var(--bg-base)', color: 'var(--text)',
      overflowY: 'auto', minHeight: 100, display: 'flex',
      flexDirection: isLandscape ? 'row' : 'column',
    }}>
      {children}
      <Toast />
    </div>
  );
};
