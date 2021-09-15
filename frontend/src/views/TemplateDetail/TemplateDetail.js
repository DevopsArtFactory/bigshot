import React from "react";
import axios from "axios";
// @material-ui/core components
import { withStyles } from "@material-ui/core/styles";

import InputLabel from "@material-ui/core/InputLabel";
// core components
import GridItem from "components/Grid/GridItem.js";
import GridContainer from "components/Grid/GridContainer.js";
import Button from "components/CustomButtons/Button.js";
import Card from "components/Card/Card.js";
import CardHeader from "components/Card/CardHeader.js";
import CardBody from "components/Card/CardBody.js";
import CardFooter from "components/Card/CardFooter.js";

import avatar from "assets/img/faces/marc.jpg";
import Loader from 'react-loader-spinner'
import PropTypes from "prop-types";
import config from "../../config"
import InfoList from "./InfoList";

const styles = {
  cardCategoryWhite: {
    color: "rgba(255,255,255,.62)",
    margin: "0",
    fontSize: "14px",
    marginTop: "0",
    marginBottom: "0"
  },
  cardTitleWhite: {
    color: "#FFFFFF",
    marginTop: "0px",
    minHeight: "auto",
    fontWeight: "300",
    fontFamily: "'Roboto', 'Helvetica', 'Arial', sans-serif",
    marginBottom: "3px",
    textDecoration: "none"
  }
};

class TemplateDetail extends React.Component {
  constructor(props) {
    super(props);

    this.state = {
      isSaved: true,
      loading: true,
      initModification: false,
      modificationTarget: null,
      modificationTargetIndex: -1,
      modificationTargetMode: 1,
      verifyResult: null,
      item: {},
    }
  }

  componentDidMount = async () =>  {
    const { params } = this.props.match;
    try {
      const template = params.template
      if (template.length > 0) {
        const response = await axios.get(
            config.api + '/detail/'+ template,
        );

        if (response.status === 200) {
          this.setState({
            item: response.data.body,
            loading: false,
          });
        }
      }
    } catch (e) {
      console.error(e);
    }
  };

  goBack = () => {
    this.props.history.goBack()
  }

  render(){
    const { classes } = this.props
    if (this.state.loading) {
      return (
          <div className="content">
            <Loader
                type="ThreeDots"
                color="#00BFFF"
                height={100}
                width={100}
                timeout={0} //3 secs
            />
          </div>
      )
    } else {
      return (
          <div>
            <GridContainer>
              <GridItem xs={12} sm={12} md={8}>
                <Card>
                  <CardHeader color="primary">
                    <h4 className={classes.cardTitleWhite}>Template Details</h4>
                    <p className={classes.cardCategoryWhite}>Detail information about template</p>
                  </CardHeader>
                  <CardBody>
                    <GridContainer>
                      <GridItem xs={12} sm={12}>
                        <InfoList
                            item={this.state.item}
                        />
                      </GridItem>
                    </GridContainer>
                  </CardBody>
                  <CardFooter>
                    <Button color="primary">Save</Button>
                  </CardFooter>
                </Card>
              </GridItem>
            </GridContainer>
          </div>
      );
    }
  }
}

TemplateDetail.propTypes = {
  classes: PropTypes.object.isRequired,
};

export default withStyles(styles)(TemplateDetail)
