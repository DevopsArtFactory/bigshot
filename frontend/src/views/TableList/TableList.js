import React from "react";
import axios from "axios";

// @material-ui/core components
import { withStyles } from "@material-ui/core/styles";
import PropTypes from 'prop-types';

// core components
import GridItem from "components/Grid/GridItem.js";
import GridContainer from "components/Grid/GridContainer.js";
import Table from "components/Table/Table.js";
import Card from "components/Card/Card.js";
import CardHeader from "components/Card/CardHeader.js";
import CardBody from "components/Card/CardBody.js";
import config from "../../config";

const styles = {
  cardCategoryWhite: {
    "&,& a,& a:hover,& a:focus": {
      color: "rgba(255,255,255,.62)",
      margin: "0",
      fontSize: "14px",
      marginTop: "0",
      marginBottom: "0"
    },
    "& a,& a:hover,& a:focus": {
      color: "#FFFFFF"
    }
  },
  cardTitleWhite: {
    color: "#FFFFFF",
    marginTop: "0px",
    minHeight: "auto",
    fontWeight: "300",
    fontFamily: "'Roboto', 'Helvetica', 'Arial', sans-serif",
    marginBottom: "3px",
    textDecoration: "none",
    "& small": {
      color: "#777",
      fontSize: "65%",
      fontWeight: "400",
      lineHeight: "1"
    }
  }
};

class TableList extends React.Component {
  constructor(props) {
    super(props);

    this.state = {
      loading: true,
      items: [],
    }
  }

  componentDidMount = async () =>  {
    try {
      const response = await axios.get(
          config.api + '/list'
      );

      if (response.status === 200) {
        this.setState({
          items: response.data.body,
          loading: false,
        });
      }

    } catch (e) {
      console.error(e);
    }
  };

  createHeader = () => {
    const data = ["Name", "# of regions", "# of targets", "interval"];

    return data
  }

  createData = () => {
    const data = [];

    this.state.items.forEach((e) => {
      data.push([e.name, e.regions.length.toString(), e.targets.length.toString(), e.interval.toString()])
    })

    return data
  }



  render() {
    const { classes } = this.props
    return (
        <GridContainer>
          <GridItem xs={12} sm={12} md={12}>
            <Card>
              <CardHeader color="primary">
                <h4 className={classes.cardTitleWhite}>Bigshot synthetic list</h4>
                <p className={classes.cardCategoryWhite}>
                  Here's the list of current synthetic list
                </p>
              </CardHeader>
              <CardBody>
                <Table
                    tableHeaderColor="primary"
                    tableHead={this.createHeader()}
                    tableData={this.createData()}
                />
              </CardBody>
            </Card>
          </GridItem>
        </GridContainer>
    );
  }
}

TableList.propTypes = {
  classes: PropTypes.object.isRequired,
};

export default withStyles(styles)(TableList)