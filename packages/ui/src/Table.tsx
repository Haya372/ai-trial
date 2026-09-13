import { cn } from 'cn'

type TableProps = Omit<React.HTMLAttributes<HTMLTableElement>, 'className'>

function Table({ ...props }: TableProps) {
  return (
    <div className={cn('w-full overflow-auto')}>
      <table
        data-slot="table"
        className={cn('w-full caption-bottom text-sm')}
        {...props}
      />
    </div>
  )
}

type TableHeaderProps = Omit<
  React.HTMLAttributes<HTMLTableSectionElement>,
  'className'
>

function TableHeader({ ...props }: TableHeaderProps) {
  return (
    <thead
      data-slot="table-header"
      className={cn('[&_tr]:border-b')}
      {...props}
    />
  )
}

type TableBodyProps = Omit<
  React.HTMLAttributes<HTMLTableSectionElement>,
  'className'
>

function TableBody({ ...props }: TableBodyProps) {
  return (
    <tbody
      data-slot="table-body"
      className={cn('[&_tr:last-child]:border-0')}
      {...props}
    />
  )
}

type TableRowProps = Omit<
  React.HTMLAttributes<HTMLTableRowElement>,
  'className'
>

function TableRow({ ...props }: TableRowProps) {
  return (
    <tr
      data-slot="table-row"
      className={cn(
        'border-b transition-colors hover:bg-muted/50 data-[state=selected]:bg-muted',
      )}
      {...props}
    />
  )
}

type TableHeadProps = Omit<
  React.ThHTMLAttributes<HTMLTableCellElement>,
  'className'
>

function TableHead({ ...props }: TableHeadProps) {
  return (
    <th
      data-slot="table-head"
      className={cn(
        'h-12 px-4 text-left align-middle text-sm font-medium text-muted-foreground [&:has([role=checkbox])]:pr-0',
      )}
      {...props}
    />
  )
}

type TableCellProps = Omit<
  React.TdHTMLAttributes<HTMLTableCellElement>,
  'className'
>

function TableCell({ ...props }: TableCellProps) {
  return (
    <td
      data-slot="table-cell"
      className={cn('p-4 align-middle [&:has([role=checkbox])]:pr-0')}
      {...props}
    />
  )
}

export { Table, TableHeader, TableBody, TableRow, TableHead, TableCell }
