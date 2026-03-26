import React from 'react';
import { QRCodeSVG } from 'qrcode.react';
import { FaTimesCircle } from 'react-icons/fa';
import { Toast } from './UI';

export const ViewExitButton: React.FC<{ onClick: () => void }> = ({ onClick }) => (
  <button
    style={{ padding: 10, fontSize: 50, position: 'absolute', color: 'var(--text)', right: 0, opacity: 0, transition: 'opacity .3s ease-out' }}
    onMouseEnter={(e) => (e.currentTarget.style.opacity = '1')}
    onMouseLeave={(e) => (e.currentTarget.style.opacity = '0')}
    onClick={(e) => {
      if (window.confirm('Are you sure want to exit?')) onClick();
      else e.preventDefault();
    }}
  >
    <FaTimesCircle />
  </button>
);

export const ViewTextContent: React.FC<React.PropsWithChildren<{ className?: string; style?: React.CSSProperties }>> = ({
  className,
  children,
  style,
}) => (
  <div className={className}>
    <p style={{ fontSize: 60, color: 'var(--text)', textAlign: 'center', padding: 16, whiteSpace: 'pre-wrap', ...style }}>{children}</p>
  </div>
);

export const generateClientUrl = (path: string) =>
  window.location.protocol + '//' + window.location.host + path;

export const QRCodeContent: React.FC<React.PropsWithChildren<{
  className?: string;
  value: string;
}>> = ({ className, children, value }) => (
  <div className={className}>
    <p style={{ fontSize: 60, color: 'var(--text)', textAlign: 'center', padding: 16, whiteSpace: 'pre-wrap' }}>{children}</p>
    <QRCodeSVG style={{ height: '65vh', width: '65vh', display: 'block', margin: '0 auto' }} size={2000} marginSize={4} bgColor="#fff" value={value} />
  </div>
);

export const ViewBlockContent: React.FC<React.PropsWithChildren> = ({ children }) => {
  const items = React.Children.toArray(children);
  return (
    <div style={{ fontSize: 60, color: 'var(--text)' }}>
      {items.map((child, i) => (
        <div key={i} style={{ padding: 10, ...(i < items.length - 1 ? { borderBottom: '10px solid var(--bg-dark)' } : {}) }}>
          {child}
        </div>
      ))}
    </div>
  );
};

export const ViewContent: React.FC<React.PropsWithChildren> = ({ children }) => (
  <div style={{ width: '100%', height: '100%', display: 'flex', justifyContent: 'center', alignItems: 'center' }}>{children}</div>
);

export const GameView: React.FC<React.PropsWithChildren> = ({ children }) => (
  <div style={{ height: '100%', backgroundColor: 'var(--bg-base)', display: 'flex', flexDirection: 'column' }}>
    {children}
    <Toast />
  </div>
);
