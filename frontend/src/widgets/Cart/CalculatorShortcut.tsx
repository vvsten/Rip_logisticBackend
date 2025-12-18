import { useEffect, useState } from 'react';

/**
 * Кнопка перехода к расчёту перевозки / черновику заявки
 * Справа вверху, с бэйджем количества
 */
export function CalculatorShortcut() {
  const [count, setCount] = useState<number>(0);
  const [logisticRequestId, setLogisticRequestId] = useState<number | null>(null);

  useEffect(() => {
    const load = async () => {
      try {
        const res = await fetch('/api/logistic-requests/draft');
        if (res.ok) {
          const data = await res.json();
          const c = typeof data?.count === 'number' ? data.count : 0;
          const id = data?.draft_logistic_request?.id ?? null;
          setCount(c);
          setLogisticRequestId(id);
          return;
        }
      } catch {}
      try {
        const res2 = await fetch('/api/logistic-requests/draft/count');
        if (res2.ok) {
          const data2 = await res2.json();
          setCount(typeof data2?.count === 'number' ? data2.count : 0);
        }
      } catch {}
    };
    load();
  }, []);

  const href = logisticRequestId ? `/delivery-quote?request_id=${logisticRequestId}` : '/delivery-quote';
  const isDisabled = count <= 0;

  return (
    <div className="calculator-shortcut">
      {isDisabled ? (
        <a className="calculator-btn is-disabled" aria-disabled="true">
          🧮 Расчёт перевозки
          <span className="cart-count" id="cartCount">{count || ''}</span>
        </a>
      ) : (
        <a href={href} className="calculator-btn" style={{ textDecoration: 'none' }}>
          🧮 Расчёт перевозки
          <span className="cart-count" id="cartCount">{count}</span>
        </a>
      )}
    </div>
  );
}


