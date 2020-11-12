import React from "react";

// reactstrap components
import {
  Card,
  CardHeader,
  CardBody,
  Row,
  Col,
  FormGroup,
  Label,
  Input,
  Button,
  CardFooter,
    Table,
} from "reactstrap";
import axios from "axios";

class TemplateDetails extends React.Component {
  constructor(props) {
    super(props);

    this.state = {
      loading: true,
      item: {},
    }

    this.Reset = this.Reset.bind(this)
    this.Save = this.Save.bind(this)
  }

  componentDidMount = async () =>  {
    const { params } = this.props.match;
    try {
      const template = params.template
      if (template.length > 0) {
        const response = await axios.get(
            'http://localhost:8765/detail/'+ template,
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

  Save = () => {
    console.log(this.state.item)
  }

  HandleInput = (identifier) => {
    console.log(this.state.item)
  }

  Reset = () => {
  }

  render() {
    const { item } = this.state
    console.log(item)

    if (this.state.loading) {
      return (
          <>
            <div className="content">
              <div>page is loading... </div>
            </div>
            </>
      )
    }
    return (
      <>
        <div className="content">
          <Row>
            <Col md="12">
              <Card>
                <CardHeader>
                  <Button color="secondary" size="sm" onClick={this.goBack}>Go Back</Button>
                  <h3 className="title">Detail Page</h3>
                </CardHeader>
                <CardBody>
                  <form>
                    <h1>Basic Information</h1>
                    <div className="form-row">
                      <FormGroup className="col-md-12">
                        <Label for="name_input">Name</Label>
                        <Input type="text"  id="name_input" defaultValue={item.Name}/>
                      </FormGroup>
                    </div>
                    <div className="form-row">
                      <FormGroup className="col-md-3">
                        <Label for="interval_input">Interval</Label>
                        <Input type="number"  id="interval_input" defaultValue={item.Interval}/>
                      </FormGroup>
                      <FormGroup className="col-md-3">
                        <Label for="timeout_input">Timeout</Label>
                        <Input type="number"  id="timeout_input" defaultValue={item.Timeout}/>
                      </FormGroup>
                    </div>
                    <h1>SlackURLs</h1>
                    {item.SlackURLs ?
                      <ul>
                        {item.SlackURLs.map((slack, i) => {
                          return (
                              <li key={"slack_"+i}>{slack}</li>
                          )
                        })
                        }
                      </ul> : <div>No slack urls</div>
                    }

                    <h1>Regions</h1>
                    {item.Regions ?
                        <Table responsive>
                          <thead>
                            <tr>
                              <th className="text-center">#</th>
                              <th className="text-center">Name</th>
                              <th className="text-right">Actions</th>
                            </tr>
                          </thead>
                          <tbody>
                          {item.Regions.map((region, i) => {
                            return (
                            <tr key={"region"+1}>
                              <td className="text-center">{i+1}</td>
                              <td className="text-center">{region.region}</td>
                              <td className="text-right">
                                <Button className="btn-icon btn-simple" color="danger" size="sm" onClick={() => this.Delete}>
                                  <i className="fa fa-times" />
                                </Button>{` `}
                              </td>
                            </tr>
                            )
                          })
                          }
                          </tbody>
                        </Table>  : <div>No region specified</div>
                    }

                    <h1>Targets</h1>
                    {item.Targets.map((target, i) => (
                        <div key={i}>
                          <div className="form-row">
                            <FormGroup className="col-md-3">
                            <Label for="method_input">Method</Label>
                            <Input type="select" name="select" id="method_input" value={target.method}>
                              <option value="GET">GET</option>
                              <option value="POST">POST</option>
                              <option value="PUT">PUT</option>
                            </Input>
                          </FormGroup>
                          <FormGroup className="col-md-9">
                            <Label for="url_input">URL</Label>
                            <Input type="text"  id="url_input" value={target.url}/>
                          </FormGroup>
                          </div>
                          {target.header && <blockquote>
                            <div className="blockquote blockquote-secondary">
                              <strong>Header</strong>
                              {Object.entries(target.header).map(([key,value])=>{
                                return (
                                    <div className="form-row" key={"header_"+key}>
                                      <FormGroup className="col-md-3">
                                        <Label for="key_input">Key</Label>
                                        <Input type="text"  id="key_input" value={key}/>
                                      </FormGroup>
                                      <FormGroup className="col-md-9">
                                        <Label for="value_input">Value</Label>
                                        <Input type="text"  id="value_input" value={value.toString()}/>
                                      </FormGroup>
                                    </div>
                                );
                            })}
                          </div>
                          </blockquote>}

                          {target.body && <blockquote>
                            <p className="blockquote blockquote-primary">
                              <h5>Body</h5>
                              {Object.entries(target.body).map(([key,value])=>{
                                return (
                                    <div className="form-row">
                                      <FormGroup className="col-md-3">
                                        <Label for="key_input">Key</Label>
                                        <Input type="text"  id="key_input" value={key}/>
                                      </FormGroup>
                                      <FormGroup className="col-md-9">
                                        <Label for="value_input">Value</Label>
                                        <Input type="text"  id="value_input" value={value.toString()}/>
                                      </FormGroup>
                                    </div>
                                );
                              })}
                            </p>
                          </blockquote>}
                        </div>
                    ))}
                  </form>
                </CardBody>
                <CardFooter>
                  <Button color="success" size="sm" onClick={this.Save}>Save</Button>
                  <Button color="secondary" size="sm" onClick={this.Reset}>Reset</Button>
                </CardFooter>
              </Card>
            </Col>
          </Row>
        </div>
      </>
    );
  }
}

export default TemplateDetails;