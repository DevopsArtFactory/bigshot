/*!

=========================================================
* Black Dashboard React v1.1.0
=========================================================

* Product Page: https://www.creative-tim.com/product/black-dashboard-react
* Copyright 2020 Creative Tim (https://www.creative-tim.com)
* Licensed under MIT (https://github.com/creativetimofficial/black-dashboard-react/blob/master/LICENSE.md)

* Coded by Creative Tim

=========================================================

* The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

*/
import React, { useState, useEffect } from "react";
import axios from 'axios';

// reactstrap components
import {
  Table,
  Button
} from "reactstrap";

class SystheticList extends React.Component {
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
          'http://localhost:8765/list'
      );

      if (response.status == 200) {
        this.setState({
          items: response.data.body,
          loading: false,
        });
      }

    } catch (e) {
      console.error(e);
    }
  };

  showDetails = (template) => {
    this.props.history.push('/admin/detail/'+template)
  }

  render(){
    return (
        <>
          <div className="content">
            <Table>
              <thead>
                <tr>
                  <th className="text-center">#</th>
                  <th>Name</th>
                  <th>Regions</th>
                  <th>Targets</th>
                  <th className="text-center">Interval</th>
                  <th className="text-center">Timeout</th>
                  <th className="text-right">Actions</th>
                </tr>
              </thead>
              <tbody>
                {this.state.items.map((item, i) => (
                    <tr key={"id_"+i}>
                      <td className="text-center">{i+1}</td>
                      <td>{ item.Name }</td>
                      <td>{ item.Regions.length }</td>
                      <td>{ item.Targets.length }</td>
                      <td className="text-center">{ item.Interval }</td>
                      <td className="text-center">{ item.Timeout }</td>
                      <td className="text-right">
                      <Button className="btn-icon btn-round" color="success" size="sm" onClick={() => this.showDetails(item.Name)}>
                      <i className="tim-icons icon-align-left-2"></i>
                      </Button>{` `}
                      <Button className="btn-icon btn-round" color="danger" size="sm">
                      <i className="fa fa-times" />
                      </Button>{` `}
                      </td>
                    </tr>
                ))}
              </tbody>
            </Table>
          </div>
        </>
    );
  }
}

export default SystheticList;
