type FrontmatterTableProps = {
  frontmatter: Record<string, unknown> | null | undefined;
};

const CHIP_MAX = 20;

function renderChip(value: string, index: number): JSX.Element {
  return (
    <span key={index} className="fm-chip">
      {value}
    </span>
  );
}

function renderValue(value: unknown): JSX.Element {
  if (value === null || value === undefined) {
    return <span className="fm-null">&mdash;</span>;
  }

  if (typeof value === "boolean") {
    return <code className="fm-bool">{value ? "true" : "false"}</code>;
  }

  if (typeof value === "number") {
    return <code className="fm-number">{String(value)}</code>;
  }

  if (Array.isArray(value)) {
    if (value.length === 0) {
      return <span className="fm-null">&mdash;</span>;
    }
    return (
      <span className="fm-chips">
        {value.slice(0, CHIP_MAX).map((item, i) =>
          renderChip(String(item), i)
        )}
        {value.length > CHIP_MAX ? (
          <span className="fm-chip fm-chip-more">+{value.length - CHIP_MAX}</span>
        ) : null}
      </span>
    );
  }

  if (typeof value === "object") {
    return (
      <table className="fm-nested-table">
        <tbody>
          {Object.entries(value).map(([k, v]) => (
            <tr key={k}>
              <td className="fm-nested-key">{k}</td>
              <td className="fm-nested-value">{renderValue(v)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    );
  }

  const text = String(value);
  if (/^https?:\/\/\S+$/i.test(text)) {
    return (
      <a href={text} target="_blank" rel="noopener noreferrer">
        {text}
      </a>
    );
  }

  return <span>{text}</span>;
}

export default function FrontmatterTable({ frontmatter }: FrontmatterTableProps) {
  if (!frontmatter || Object.keys(frontmatter).length === 0) {
    return null;
  }

  const entries = Object.entries(frontmatter);

  return (
    <div className="fm-container">
      <table className="fm-table">
        <tbody>
          {entries.map(([key, value]) => (
            <tr key={key}>
              <td className="fm-key">{key}</td>
              <td className="fm-value">{renderValue(value)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
