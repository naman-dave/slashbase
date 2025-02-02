import styles from './jsontable.module.scss'
import React, { useRef, useState } from 'react'
import { useRowSelect, useTable } from 'react-table'
import { DBConnection, DBQueryData } from '../../../data/models'
import JsonCell from './jsoncell'
import AddModal from './addmodel'
import apiService from '../../../network/apiService'
import toast from 'react-hot-toast'

type JsonTablePropType = {
    queryData: DBQueryData,
    dbConnection: DBConnection
    mName: string,
    isEditable: boolean,
    showHeader?: boolean,
    onAddData: (newData: any) => void,
    onDeleteRows: (indexes: number[]) => void,
    updateCellData: (underscoreId: string, newData: object) => void,
    onFilterChanged: (newFilter: string[] | undefined) => void,
    onSortChanged: (newSort: string[] | undefined) => void,
}

const JsonTable = ({ queryData, dbConnection, mName, isEditable, showHeader, onAddData, onDeleteRows, updateCellData, onFilterChanged, onSortChanged }: JsonTablePropType) => {

    const [isAdding, setIsAdding] = useState<boolean>(false)
    const [editingCellIndex, setEditingCellIndex] = useState<(number | null)>(null)

    const filterRef = useRef<HTMLInputElement>(null);
    const sortRef = useRef<HTMLInputElement>(null);

    const data = React.useMemo(
        () => queryData.data,
        [queryData]
    )

    const displayColumns = ["data"]

    const columns = React.useMemo(
        () => displayColumns.map((col, i) => ({
            Header: <>{col}</>,
            accessor: (i).toString(),
        })),
        [queryData]
    )

    const defaultColumn = {
        Cell: JsonCell,
    }

    const startEditing = (index: number | null) => {
        if (!isEditable) {
            return
        }
        setEditingCellIndex(index)
    }

    const changeFilter = () => {
        let filter: string[] | undefined = undefined
        let filterText = filterRef.current!.value.trim()
        if (filterText !== '' && filterText.startsWith("{") && filterText.endsWith("}")) {
            filter = [filterText]
        }
        onFilterChanged(filter)
    }

    const changeSort = () => {
        if (!isEditable) {
            return
        }
        let sort: string[] | undefined = undefined
        let sortText = sortRef.current!.value.trim()
        if (sortText !== '' && sortText.startsWith("{") && sortText.endsWith("}")) {
            sort = [sortText]
        }
        onSortChanged(sort)
    }


    const onSaveCell = async (underscoreId: string, newData: string) => {
        const result = await apiService.updateDBSingleData(dbConnection.id, "", mName, underscoreId, "", newData)
        if (result.success) {
            updateCellData(underscoreId, JSON.parse(newData))
            startEditing(null)
            toast.success('1 row updated');
        } else {
            toast.error(result.error!);
        }
    }

    const {
        getTableProps,
        getTableBodyProps,
        headerGroups,
        rows,
        prepareRow,
        state,
    } = useTable<any>({
        columns,
        data,
        defaultColumn,
        ...{ editingCellIndex, startEditing, onSaveCell }
    }, useRowSelect, hooks => {
        if (isEditable)
            hooks.visibleColumns.push(columns => [
                {
                    id: 'selection',
                    Header: HeaderSelectionComponent,
                    Cell: CellSelectionComponent,
                },
                ...columns,
            ]
            )
    })

    const newState: any = state // temporary typescript hack
    const selectedRows: number[] = Object.keys(newState.selectedRowIds).map(x => parseInt(x))
    const selectedUnderscoreIDs = rows.filter((_, i) => selectedRows.includes(i)).map(x => x.original['_id']).filter(x => x)

    const onDeleteBtnPressed = async () => {
        if (selectedUnderscoreIDs.length > 0) {
            const result = await apiService.deleteDBData(dbConnection.id, "", mName, selectedUnderscoreIDs)
            if (result.success) {
                toast.success('rows deleted');
                onDeleteRows(selectedRows)
            } else {
                toast.error(result.error!);
            }
        }
    }

    return (
        <React.Fragment>
            {(showHeader || isEditable) && <div className={styles.tableHeader}>
                <div className="columns">
                    <div className="column is-3">
                        <div className="field has-addons">
                            <p className="control">
                                <input ref={filterRef} className="input" type="text" placeholder="{ field: 'Value'}" />
                            </p>
                            <p className="control">
                                <button className="button" onClick={changeFilter}>Filter</button>
                            </p>
                        </div>
                    </div>
                    <div className="column is-6">
                        <div className="field has-addons">
                            <p className="control">
                                <input ref={sortRef} className="input" type="text" placeholder="{ field: 1 or -1}" />
                            </p>
                            <p className="control">
                                <button className="button" onClick={changeSort}>Sort</button>
                            </p>
                        </div>
                    </div>
                    {isEditable && <React.Fragment>
                        <div className="column is-3 is-flex is-justify-content-flex-end">
                            <button className="button" disabled={selectedUnderscoreIDs.length === 0} onClick={onDeleteBtnPressed}>
                                <span className="icon is-small">
                                    <i className="fas fa-trash" />
                                </span>
                            </button>
                            &nbsp;&nbsp;
                            <button className="button is-primary" onClick={() => { setIsAdding(true) }}>
                                <span className="icon is-small">
                                    <i className="fas fa-plus" />
                                </span>
                            </button>
                        </div>
                    </React.Fragment>}
                </div>
            </div>}
            {isAdding &&
                <AddModal
                    dbConnection={dbConnection}
                    mName={mName}
                    onClose={() => { setIsAdding(false) }}
                    onAddData={onAddData} />
            }
            <div className="table-container">
                <table {...getTableProps()} className={"table is-bordered is-striped is-narrow is-hoverable is-fullwidth"}>
                    <thead>
                        {headerGroups.map(headerGroup => (
                            <tr {...headerGroup.getHeaderGroupProps()} key={"header"}>
                                {headerGroup.headers.map(column => (
                                    <th {...column.getHeaderProps()} key={column.id}>
                                        {column.render('Header')}
                                    </th>
                                ))}
                            </tr>
                        ))}
                    </thead>
                    <tbody {...getTableBodyProps()}>
                        {rows.map((row, rowIndex) => {
                            prepareRow(row)
                            const selectedRow: any = row // temp type hack 
                            return (
                                <tr {...row.getRowProps()} key={row.id} className={selectedRow.isSelected ? 'is-selected' : ''}>
                                    {row.cells.map(cell => {
                                        return (<td {...cell.getCellProps()} onDoubleClick={() => startEditing(rowIndex)} key={row.id + "" + cell.column.id}>
                                            {cell.render('Cell')}
                                        </td>
                                        )
                                    })}
                                </tr>
                            )
                        })}
                    </tbody>
                </table>
            </div>
        </React.Fragment>
    )
}

const IndeterminateCheckbox = React.forwardRef<HTMLInputElement, { indeterminate: boolean }>(
    ({ indeterminate, ...rest }, ref) => {
        const defaultRef = React.useRef()
        const resolvedRef: any = ref || defaultRef

        React.useEffect(() => {
            resolvedRef.current.indeterminate = indeterminate
        }, [resolvedRef, indeterminate])

        return (
            <>
                <input type="checkbox" ref={resolvedRef} {...rest} />
            </>
        )
    }
)
IndeterminateCheckbox.displayName = 'IndeterminateCheckbox';

const HeaderSelectionComponent = ({ getToggleAllRowsSelectedProps }: any) => (
    <div>
        <IndeterminateCheckbox {...getToggleAllRowsSelectedProps()} />
    </div>
)

const CellSelectionComponent = ({ row }: any) => (
    <div>
        <IndeterminateCheckbox {...row.getToggleRowSelectedProps()} />
    </div>
)

export default JsonTable