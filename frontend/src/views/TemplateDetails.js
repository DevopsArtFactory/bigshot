import React from "react";
// reactstrap components
import {
  Button,
  Card,
  CardBody,
  CardFooter,
  CardHeader,
  Col,
  FormGroup,
  Input,
  Label,
  Row,
  Table,
  Modal,
  ModalBody,
  ModalHeader,
  Alert
} from "reactstrap";
import axios from "axios";
import Loader from 'react-loader-spinner'
import { toast } from 'react-toastify';

const { APP_HOST } = process.env;

class TemplateDetails extends React.Component {
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

    this.Save = this.Save.bind(this)
    this.DeleteRegion = this.DeleteRegion.bind(this)
    this.toggleTargetModal = this.toggleTargetModal.bind(this);
  }

  componentDidMount = async () =>  {
    const { params } = this.props.match;
    try {
      const template = params.template
      if (template.length > 0) {
        const response = await axios.get(
            APP_HOST + '/detail/'+ template,
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

  Save = async () => {
    this.setState((status, props) => ({
      isSaving: false,
    }));

    const { params } = this.props.match;
    try {
      const template = params.template

      const response = await axios.post(
          APP_HOST + '/save/template/'+ template,
          this.state.item,
      );

      if (response.status === 200) {
        this.setState((status, props) => ({
          isSaving: true,
        }));
      }

      toast("successfully saved", {
        position: "top-right",
        autoClose: 3000,
        hideProgressBar: false,
        closeOnClick: true,
        pauseOnHover: true,
        draggable: true,
      }
    )
    } catch (e) {
      console.error(e);
    }
  };

  DeleteRegion = (region) => {
    const item = {...this.state.item};
    item.Regions = [...item.Regions.filter(x => x.region !== region)]
    this.setState((status, props) => ({
      item: item,
    }));
  }

  AddNewRegion = (region) => {
    const item = {...this.state.item};
    const newRegion = document.getElementById("new_region_input").value;
    if (newRegion.length === 0) {
      alert("please write the region ID");
      return;
    }
    const duplicated = [...item.Regions.filter(x => x.region === newRegion)];
    if (duplicated.length > 0) {
      alert(newRegion + " is already registered")
      document.getElementById('new_region_input').value = ""
      return
    }
    item.Regions = item.Regions.concat({region: newRegion})
    this.setState((status, props) => ({
      item: item,
    }));

    document.getElementById('new_region_input').value = ""
  }

  AddNewHeader = () => {
    const item = {...this.state.modificationTarget};
    const newHeaderKey = document.getElementById("new_header_key").value;
    const newHeaderValue = document.getElementById("new_header_value").value;
    if (newHeaderKey.length === 0 || newHeaderValue === 0) {
      alert("please write both key and value");
      return;
    }
    if (item.header && item.header[newHeaderKey]) {
      alert(newHeaderKey + " already exists")
      return
    }
    if(!item.header) {
      item.header = {}
    }
    item.header[newHeaderKey] = newHeaderValue
    this.setState((status, props) => ({
      modificationTarget: item,
    }));

    document.getElementById('new_header_key').value = ""
    document.getElementById('new_header_value').value = ""
  }

  AddNewBody = () => {
    const item = {...this.state.modificationTarget};
    const newBodyKey = document.getElementById("new_body_key").value;
    const newBodyValue = document.getElementById("new_body_value").value;
    if (newBodyKey.length === 0 || newBodyValue === 0) {
      alert("please write both key and value");
      return;
    }
    if (item.body && item.body[newBodyKey]) {
      alert(newBodyKey + " already exists");
      return;
    }
    if(!item.body) {
      item.body = {};
    }
    item.body[newBodyKey] = newBodyValue;
    this.setState((status, props) => ({
      modificationTarget: item,
    }));

    document.getElementById('new_body_key').value = ""
    document.getElementById('new_body_value').value = ""
  }

  RemoveHeader = (key) => {
    const item = {...this.state.modificationTarget};
    delete item.header[key]

    if (Object.entries(item.header).length === 0) {
      item.header = null
    }
    this.setState((status, props) => ({
      modificationTarget: item,
    }));
  }

  RemoveBody = (key) => {
    const item = {...this.state.modificationTarget};
    delete item.body[key]

    if (Object.entries(item.body).length === 0) {
      item.body = null
    }
    this.setState((status, props) => ({
      modificationTarget: item,
    }));
  }

  ModifyTarget = (id) => {
    const target = {...this.state.item.Targets[id]}
    this.setState((status, props) => ({
      modificationTarget: target,
      modificationTargetIndex: id,
      modificationTargetMode: 1,
    }));

    this.toggleTargetModal()
  }

  CreateTarget = () => {
    this.setState((status, props) => ({
      modificationTarget: {},
      modificationTargetMode: 0,
    }));

    this.toggleTargetModal()
  }

  RemoveTarget = (id) => {
    const item = {...this.state.item};
    item.Targets = [...item.Targets.filter((x, i) => i !== id)]

    if (item.Targets.length === 0) {
      item.Targets = null
    }
    this.setState((status, props) => ({
      item: item,
    }));
  }

  SaveTargetModification = () => {
    const item = {...this.state.item};
    if (this.state.modificationTargetMode === 0){
      item.Targets = item.Targets.concat({...this.state.modificationTarget});
    } else {
      item.Targets[this.state.modificationTargetIndex] = {...this.state.modificationTarget};
    }

    this.setState((status, props) => ({
        item: item,
        modificationTargetIndex: -1,
        modificationTargetMode: 1,
        modificationTarget: null,
    }));

    this.toggleTargetModal();
  }

  HandleTargetModificationInput = (event, id) => {
    const item = {...this.state.modificationTarget};
    item[id] = event.target.value;

    this.setState((status, props) => ({
      modificationTarget: item,
    }));
  }

  Verify = async () => {
    const item = {...this.state.modificationTarget};
    try {
      const response = await axios.post(
          APP_HOST + '/verify-target',
          item,
      );

      if (response.status === 200) {
        this.setState((status, props) => ({
          verifyResult: {
            response: response.data.body["Response"],
          },
        }));
      }
    } catch (e) {
      alert(e);
    }
  }

  toggleTargetModal(){
    this.setState({
      initModification: !this.state.initModification,
      verifyResult: null,
    });
  }

  render() {
    const { item } = this.state

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
      return (<>
            <div className="content"><Row> <Col md="12"> <Card> <CardHeader><h1 className="title">Detail Page</h1>
            </CardHeader> <CardBody>
              <form><h1>Basic Information</h1>
                <div className="form-row"><FormGroup className="col-md-12"> <Label for="name_input">Name</Label> <Input
                    type="text" id="name_input" defaultValue={item.Name} disabled/> </FormGroup></div>
                <div className="form-row"><FormGroup className="col-md-3"> <Label for="interval_input">Interval</Label>
                  <Input type="number" id="interval_input" defaultValue={item.Interval}/> </FormGroup> <FormGroup
                    className="col-md-3"> <Label for="timeout_input">Timeout</Label> <Input type="number"
                                                                                            id="timeout_input"
                                                                                            defaultValue={item.Timeout}/>
                </FormGroup></div>
                <h1>SlackURLs</h1> {item.SlackURLs ? <ul> {item.SlackURLs.map((slack, i) => {
                  return (<li key={"slack_" + i}>{slack}</li>)
                })} </ul> : <div>No slack urls</div>} <h1>Regions</h1>
                <div className="form-row"><FormGroup className="col-md-4 col-sm-12"> <Label
                    for="new_region_input">Region</Label> <Input type="text" id="new_region_input"/> </FormGroup>
                  <FormGroup className="col-md-4 col-sm-12"> <Label for="new_region_add_action">Action</Label> <Button
                      color="success" className="animation-on-hover" style={{display: 'block'}} size="sm"
                      onClick={this.AddNewRegion}> Add </Button> </FormGroup></div>
                {item.Regions ? <Table responsive>
                  <thead>
                  <tr>
                    <th className="text-center">#</th>
                    <th className="text-center">Name</th>
                    <th className="text-right">Actions</th>
                  </tr>
                  </thead>
                  <tbody> {item.Regions.map((region, i) => {
                    return (<tr key={"region_" + i}>
                      <td className="text-center">{i + 1}</td>
                      <td className="text-center">{region.region}</td>
                      <td className="text-right"><Button className="btn-icon btn-simple" color="danger" size="sm"
                                                         onClick={() => this.DeleteRegion(region.region)}> <i
                          className="fa fa-times"/> </Button>{` `} </td>
                    </tr>)
                  })} </tbody>
                </Table> : <div>No region specified</div>} <h1>Targets</h1> <Button className="btn-round"
                                                                                    color="primary"
                                                                                    onClick={this.CreateTarget}> <i
                    className="tim-icons icon-heart-2"/>{" "} Register new target </Button> {item.Targets ?
                    <Table responsive>
                      <thead>
                      <tr>
                        <th className="text-center">Method</th>
                        <th className="text-center">URL</th>
                        <th className="text-right">Action</th>
                      </tr>
                      </thead>
                      <tbody> {item.Targets.map((target, i) => (<tr key={"target" + i}>
                        <td className="text-center">{target.method}</td>
                        <td className="text-center">{target.url}</td>
                        <td className="text-right"><Button className="btn-icon btn-simple" color="success" size="sm"
                                                           onClick={() => this.ModifyTarget(i)}> <i
                            className="fa fa-edit"/> </Button>{` `} <Button className="btn-icon btn-simple"
                                                                            color="danger" size="sm"
                                                                            onClick={() => this.RemoveTarget(i)}> <i
                            className="fa fa-times"/> </Button>{` `} </td>
                      </tr>))} </tbody>
                    </Table> : <p>No target registered</p>
                }
              </form>
            </CardBody>
              <CardFooter>
                <Button color="secondary" size="lg" onClick={this.goBack}>Go Back</Button>
                <Button color="success" size="lg" onClick={this.Save}>Save</Button>
              </CardFooter>
            </Card>
            </Col>
            </Row>
            </div>
            <Modal isOpen={this.state.initModification} toggle={this.toggleTargetModal} size="xl">
              <ModalHeader className="justify-content-center" toggle={this.toggleTargetModal}>
                {this.state.modificationTargetMode ? "Target Modification" : "Create a new target"}
              </ModalHeader>
              <ModalBody>
                {this.state.modificationTarget &&
                <Card>
                  <CardBody>
                    <form>
                      <FormGroup>
                        <Label for="modify_method">Method</Label>
                        <Input
                            type="text"
                            name="modify_method"
                            id="modify_method"
                            value={this.state.modificationTarget.method}
                            onChange={(e) => this.HandleTargetModificationInput(e, 'method')}
                        />
                      </FormGroup>
                      <FormGroup>
                        <Label for="modify_url">URL</Label>
                        <Input
                            type="text"
                            name="modify_url"
                            id="modify_url"
                            autoComplete="off"
                            onChange={(e) => this.HandleTargetModificationInput(e, 'url')}
                            value={this.state.modificationTarget.url}
                        />
                      </FormGroup>
                      <div>
                        <h3><strong>Header</strong></h3>
                      </div>
                      <div className="form-row">
                        <FormGroup className="col-md-4">
                          <Label for="new_header_key">Key</Label>
                          <Input type="text" id="new_header_key"/>
                        </FormGroup>
                        <FormGroup className="col-md-6">
                          <Label for="new_header_value">Value</Label>
                          <Input type="text" id="new_header_value"/>
                        </FormGroup>
                        <FormGroup className="col-md-2">
                          <Label for="">ACTION</Label>
                          <Button color="success" className="animation-on-hover" style={{display: 'block'}} size="sm"
                                  onClick={this.AddNewHeader}>
                            Add
                          </Button>
                        </FormGroup>
                      </div>
                      {this.state.modificationTarget.header &&
                      <Table responsive>
                        <thead>
                        <tr>
                          <th className="text-left">KEY</th>
                          <th className="text-left">VALUE</th>
                          <th className="text-right">ACTION</th>
                        </tr>
                        </thead>
                        <tbody>
                        {Object.entries(this.state.modificationTarget.header).map(([key, value], i) => (
                            <tr key={"header" + i}>
                              <td className="text-lefth">{key}</td>
                              <td className="text-left">{value}</td>
                              <td className="text-right">
                                <Button className="btn-icon btn-simple" color="danger" size="sm"
                                        onClick={() => this.RemoveHeader(key)}>
                                  <i className="fa fa-times"/>
                                </Button>{` `}
                              </td>
                            </tr>
                        ))}
                        </tbody>
                      </Table>
                      }

                      <div>
                        <h3><strong>Body</strong></h3>
                      </div>
                      <div className="form-row">
                        <FormGroup className="col-md-4">
                          <Label for="new_body_key">Key</Label>
                          <Input type="text" id="new_body_key"/>
                        </FormGroup>
                        <FormGroup className="col-md-6">
                          <Label for="new_body_value">Value</Label>
                          <Input type="text" id="new_body_value"/>
                        </FormGroup>
                        <FormGroup className="col-md-2">
                          <Label for="">ACTION</Label>
                          <Button color="success" className="animation-on-hover" style={{display: 'block'}} size="sm"
                                  onClick={this.AddNewBody}>
                            Add
                          </Button>
                        </FormGroup>
                      </div>
                      {this.state.modificationTarget.body &&
                      <Table responsive>
                        <thead>
                        <tr>
                          <th className="text-left">KEY</th>
                          <th className="text-left">VALUE</th>
                          <th className="text-right">ACTION</th>
                        </tr>
                        </thead>
                        <tbody>
                        {Object.entries(this.state.modificationTarget.body).map(([key, value], i) => (
                            <tr key={"body" + i}>
                              <td className="text-lefth">{key}</td>
                              <td className="text-left">{value}</td>
                              <td className="text-right">
                                <Button className="btn-icon btn-simple" color="danger" size="sm"
                                        onClick={() => this.RemoveBody(key)}>
                                  <i className="fa fa-times"/>
                                </Button>{` `}
                              </td>
                            </tr>
                        ))}
                        </tbody>
                      </Table>
                      }
                      {this.state.verifyResult && (this.state.verifyResult.response.StatusCode == 200 ?
                          <Alert color="success">
                            {this.state.verifyResult.response.StatusCode} {this.state.verifyResult.response.StatusMsg}
                          </Alert> :
                          <Alert color="danger">
                            {this.state.verifyResult.response.StatusCode} {this.state.verifyResult.response.StatusMsg}
                          </Alert>)
                      }
                      <Button color="success" type="button" onClick={this.Verify}>
                        Verify
                      </Button>
                      <Button color="primary" type="button" onClick={this.SaveTargetModification}>
                        Save Modification
                      </Button>
                    </form>
                  </CardBody>
                </Card>
                }
              </ModalBody>
            </Modal>
          </>
      );
    }
  }
}

export default TemplateDetails;
