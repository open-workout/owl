export default function TopBar() {
  return (
    <header className="top-bar">
      <span className="top-bar-title">OWL Playground</span>
      <div className="top-bar-actions">
        <button type="button" disabled title="Coming soon">
          Publish
        </button>
        <button type="button" disabled title="Coming soon">
          Send to Phone
        </button>
      </div>
    </header>
  );
}
