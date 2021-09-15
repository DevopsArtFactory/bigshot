import React from "react";
import axios from "axios";

// core components
import GridItem from "components/Grid/GridItem.js";
import GridContainer from "components/Grid/GridContainer.js";
import Card from "components/Card/Card.js";
import CardActions from '@material-ui/core/CardActions';
import TextField from '@material-ui/core/TextField';
import Button from '@material-ui/core/Button';
import CardBody from "../../components/Card/CardBody";
import Result from "./Result";
import config from "../../config";
import PropTypes from "prop-types";
import {withStyles} from "@material-ui/core/styles";
import InputLabel from '@material-ui/core/InputLabel';
import MenuItem from '@material-ui/core/MenuItem';
import FormControl from '@material-ui/core/FormControl';
import Select from '@material-ui/core/Select';

const styles = theme =>({
    formControl: {
        margin: theme.spacing(1),
        minWidth: 120,
    },
    selectEmpty: {
        marginTop: theme.spacing(2),
    },
});

class Dashboard extends React.Component{
    constructor(props) {
        super(props);
        this.state = {
            isRun: false,
            result: {},
            url:"",
            isError: false,
            errorMsg: null,
            isResultShown: false,
            target: {
                url: "",
                method: "GET",
                port: "",
                protocol: ""
            }
        }
    }

    shoot = async () => {
        const { target } = this.state;
        try {
            const response = await axios.post(
                config.api +  '/verify-target',
                target
            );


            if (response.status === 200) {
                const data = response.data.body;

                this.setState({
                    result: data,
                    isResultShown: true,
                });
            }
        } catch (e) {
            console.error(e);
        }
    }

    parseUrl = (url) => {
        const target = {}
        let isError = false;
        let errorMsg = null;
        try {
            const parsed = new URL(url);
            target.protocol = parsed.protocol.substr(0, parsed.protocol.length-1)
            target.url = parsed.host + parsed.pathname;
            target.port = parsed.port;

            if (parsed.port == "") {
                if (target.protocol === "https") {
                    target.port = "443";
                } else {
                    target.port = "80";
                }
            }
        } catch (e) {
            isError = true;
            errorMsg = e.toString();
        }

        this.setState({
            isError: isError,
            errorMsg: errorMsg,
        });

        return target;
    }

    ChangeTargetInfo = (event, identifier) => {
        let item = null;
        if (identifier === "url")  {
            const url = event.target.value;
            const current = {...this.state.target };
            item = this.parseUrl(url);
            item["method"] = current.method
            this.setState({
                target: item,
                url: url,
            });
        } else {
            item = {...this.state.target };
            item[identifier] = event.target.value;
            this.setState({
                target: item,
            });
        }
    }

    render() {
        const { target, url, isError, errorMsg, result, isResultShown } = this.state;
        const { classes } = this.props;

        return (
            <div>
                <GridContainer>
                    <GridItem xs={12} sm={12} md={12}>
                        <Card>
                            <CardBody >
                                <FormControl className={classes.formControl}>
                                    <InputLabel id="method">Method</InputLabel>
                                    <Select
                                        labelId="method"
                                        id="method"
                                        value={target.method}
                                        onChange={(e) => this.ChangeTargetInfo(e, 'method')}
                                    >
                                        <MenuItem value={"GET"}>GET</MenuItem>
                                        <MenuItem value={"POST"}>POST</MenuItem>
                                    </Select>
                                </FormControl>
                                <TextField
                                    id="url"
                                    label="URL"
                                    placeholder="https://example.com"
                                    helperText="Enter URL"
                                    margin="normal"
                                    fullWidth
                                    value={url}
                                    error={isError}
                                    helperText={isError ? errorMsg : ""}
                                    onChange={(e) => this.ChangeTargetInfo(e, 'url')}
                                    InputLabelProps={{
                                        shrink: true,
                                    }}
                                />
                            </CardBody>
                            <CardActions>
                                <Button size="large" color="secondary" onClick={this.shoot}>
                                    Shoot
                                </Button>
                            </CardActions>
                        </Card>
                    </GridItem>

                    {/* Result */}
                    {isResultShown &&
                    <GridItem xs={12} sm={12} md={12}>
                        <Result
                            result={result}
                        />
                    </GridItem>
                    }
                </GridContainer>
            </div>
        );
    }
}

Dashboard.propTypes = {
    classes: PropTypes.object.isRequired,
};

export default withStyles(styles)(Dashboard)
