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
import Dashboard from "views/Dashboard.js";
import Icons from "views/Icons.js";
import Notifications from "views/Notifications.js";
import SystheticList from "views/SystheticList.js";
import Typography from "views/Typography.js";
import TemplateDetails from "views/TemplateDetails.js";

var routes = [
  {
    path: "/dashboard",
    name: "Dashboard",
    icon: "tim-icons icon-chart-pie-36",
    component: Dashboard,
    layout: "/admin",
  },
  {
    path: "/detail/:template",
    name: "Template details",
    icon: "tim-icons icon-single-02",
    component: TemplateDetails,
    layout: "/admin",
    skip: true,
  },
  {
    path: "/list",
    name: "Synthetic List",
    icon: "tim-icons icon-puzzle-10",
    component: SystheticList,
    layout: "/admin"
  },
  // {
  //   path: "/notifications",
  //   name: "Notifications",
  //   icon: "tim-icons icon-bell-55",
  //   component: Notifications,
  //   layout: "/admin"
  // },
  // {
  //   path: "/icons",
  //   name: "Icons",
  //   icon: "tim-icons icon-atom",
  //   component: Icons,
  //   layout: "/admin"
  // },
  // {
  //   path: "/typography",
  //   name: "Typography",
  //   icon: "tim-icons icon-align-center",
  //   component: Typography,
  //   layout: "/admin"
  // },
];
export default routes;
