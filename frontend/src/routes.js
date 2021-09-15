/*!

=========================================================
* Material Dashboard React - v1.9.0
=========================================================

* Product Page: https://www.creative-tim.com/product/material-dashboard-react
* Copyright 2020 Creative Tim (https://www.creative-tim.com)
* Licensed under MIT (https://github.com/creativetimofficial/material-dashboard-react/blob/master/LICENSE.md)

* Coded by Creative Tim

=========================================================

* The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

*/
// @material-ui/icons
import Dashboard from "@material-ui/icons/Dashboard";
// core components/views for Admin layout
import DashboardPage from "views/Dashboard/Dashboard.js";
import TemplateDetail from "views/TemplateDetail/TemplateDetail.js";
import TableList from "views/TableList/TableList.js";

const dashboardRoutes = [
  {
    path: "/main",
    name: "Main",
    icon: Dashboard,
    component: DashboardPage,
    layout: "/dashboard"
  },
  {
    path: "/list",
    name: "Bigshot List",
    icon: "content_paste",
    component: TableList,
    layout: "/dashboard"
  },
  {
    path: "/detail/:template",
    name: "Template details",
    icon: "content_paste",
    component: TemplateDetail,
    layout: "/dashboard",
    skip: true,
  },
];

export default dashboardRoutes;
