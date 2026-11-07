type SkeletonRowsProps = {
  rows?: number
  columns: number
}

export default function SkeletonRows({ rows = 5, columns }: SkeletonRowsProps) {
  return (
    <>
      {Array.from({ length: rows }, (_, row) => (
        <tr key={row} aria-hidden="true">
          {Array.from({ length: columns }, (_, column) => (
            <td key={column}>
              <span className="skeleton" />
            </td>
          ))}
        </tr>
      ))}
    </>
  )
}
