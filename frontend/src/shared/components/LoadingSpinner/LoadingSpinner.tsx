/**
 * Компонент анимации загрузки.
 * Используется при выполнении асинхронных запросов (thunk/axios).
 */

interface LoadingSpinnerProps {
  size?: 'small' | 'medium' | 'large';
  text?: string;
}

export function LoadingSpinner({ size = 'medium', text = 'Загрузка...' }: LoadingSpinnerProps) {
  const sizeMap = {
    small: '30px',
    medium: '40px',
    large: '60px',
  };

  return (
    <div style={{
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      padding: '2rem',
    }}>
      <div
        style={{
          border: '3px solid #f3f3f3',
          borderTop: '3px solid #0d6efd',
          borderRadius: '50%',
          width: sizeMap[size],
          height: sizeMap[size],
          animation: 'spin 1s linear infinite',
        }}
      />
      {text && (
        <p style={{ marginTop: '1rem', color: '#6c757d' }}>{text}</p>
      )}
      <style>{`
        @keyframes spin {
          0% { transform: rotate(0deg); }
          100% { transform: rotate(360deg); }
        }
      `}</style>
    </div>
  );
}
