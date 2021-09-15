import React from 'react';
import { makeStyles } from '@material-ui/core/styles';
import Card from "../../components/Card/Card";
import CardBody from "../../components/Card/CardBody";
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableContainer from '@material-ui/core/TableContainer';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';
import Paper from '@material-ui/core/Paper';

const useStyles = makeStyles((theme) => ({
    root: {
        width: '100%',
        maxWidth: 360,
        backgroundColor: theme.palette.background.paper,
    },
}));

function createData(name, value) {
    return { name, value };
}

function createStatusData(rawData) {
    const row = [];
    row.push(createData("Status", rawData.StatusCode + " " + rawData.StatusMsg));
    return row;
}

function createTracingData(rawData) {
    const row = [
        createData("IP Address", rawData.ConnectAddr),
        createData("DNS Lookup", rawData.DNSLookupStr),
        createData("TLS HandShaking", rawData.TLSHandShakingStr),
        createData("TCP Connection", rawData.TCPConnectionStr),
        createData("Server Processing", rawData.ServerProcessingStr),
        createData("Content Transfer", rawData.ContentTransferStr),
    ];

    return row;
}

export default function Result({ result }) {
    const classes = useStyles();
    const statusData = result['Response'];
    const tracingData = result['TracingData'];

    const rows = createStatusData(statusData)
    rows.push(...createTracingData(tracingData))

    return (
        <TableContainer component={Paper}>
            <Table className={classes.table} aria-label="simple table">
                <TableHead>
                    <TableRow>
                        <TableCell colSpan={2}>Result</TableCell>
                    </TableRow>
                </TableHead>
                <TableBody>
                    {rows.map((row, index) => (
                        <TableRow key={index}>
                            <TableCell align="center" component="th" scope="row">
                                {row.name}
                            </TableCell>
                            <TableCell align="center">{row.value}</TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </TableContainer>
    );
}
