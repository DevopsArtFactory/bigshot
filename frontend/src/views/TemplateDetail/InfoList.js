import React from 'react';
import { makeStyles } from '@material-ui/core/styles';
import IconButton from '@material-ui/core/IconButton';
import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import ListItemIcon from '@material-ui/core/ListItemIcon';
import ListItemText from '@material-ui/core/ListItemText';
import ListItemSecondaryAction from '@material-ui/core/ListItemSecondaryAction';
import Collapse from '@material-ui/core/Collapse';
import InboxIcon from '@material-ui/icons/MoveToInbox';
import PublicIcon from '@material-ui/icons/Public';
import TrackChangesIcon from '@material-ui/icons/TrackChanges';
import AccessTimeIcon from '@material-ui/icons/AccessTime';
import Forward5Icon from '@material-ui/icons/Forward5';
import SendIcon from '@material-ui/icons/Send';
import ExpandLess from '@material-ui/icons/ExpandLess';
import ExpandMore from '@material-ui/icons/ExpandMore';
import StarBorder from '@material-ui/icons/StarBorder';
import DeleteIcon from '@material-ui/icons/Delete';

const useStyles = makeStyles((theme) => ({
    root: {
        width: '100%',
        backgroundColor: theme.palette.background.paper,
    },
    nested: {
        paddingLeft: theme.spacing(4),
    },
}));

export default function InfoList({ item }) {
    const classes = useStyles();
    const [open, setOpen] = React.useState(new Map());
    const [targetOpen, setTargetOpen] = React.useState(true);

    const handleClick = (identifier) => {
        const value = !open.get(identifier);
        setOpen((open) => new Map(open).set(identifier, value))
    };

    return (
        <List
            component="nav"
            aria-labelledby="nested-list-subheader"
            className={classes.root}
        >
            <ListItem button>
                <ListItemIcon>
                    <SendIcon />
                </ListItemIcon> <ListItemText primary={item.name} />
            </ListItem>
            <ListItem button>
                <ListItemIcon>
                    <AccessTimeIcon />
                </ListItemIcon>
                <ListItemText primary={"Timeout: " + item.timeout} />
            </ListItem>
            <ListItem button>
                <ListItemIcon>
                    <Forward5Icon />
                </ListItemIcon>
                <ListItemText primary={"Interval: " + item.interval} />
            </ListItem>
            <ListItem button onClick={()=>handleClick('region')}>
                <ListItemIcon>
                    <PublicIcon />
                </ListItemIcon>
                <ListItemText primary={"Regions ("+item.regions.length+")"} />
                {open.get('region') ? <ExpandLess /> : <ExpandMore />}
            </ListItem>
            {item.regions.length > 0 &&
            <Collapse in={open.get('region')} timeout="auto" unmountOnExit>
                <List component="div" disablePadding>
                    {item.regions.map((region, i) => {
                        return (
                            <ListItem button className={classes.nested} key={"region_"+i}>
                                <ListItemIcon>
                                    <StarBorder/>
                                </ListItemIcon>
                                <ListItemText primary={region.region}/>
                                <ListItemSecondaryAction>
                                    <IconButton edge="end" aria-label="delete">
                                        <DeleteIcon />
                                    </IconButton>
                                </ListItemSecondaryAction>
                            </ListItem>
                        )
                    })}
                </List>
            </Collapse>
            }
            <ListItem button onClick={() => handleClick('target')}>
                <ListItemIcon>
                    <TrackChangesIcon />
                </ListItemIcon>
                <ListItemText primary={"Targets ("+item.targets.length+")"} />
                {open.get('target') ? <ExpandLess /> : <ExpandMore />}
            </ListItem>
            {item.targets.length > 0 &&
            <Collapse in={open.get('target')} timeout="auto" unmountOnExit>
                <List component="div" disablePadding>
                    {item.targets.map((target, i) => {
                        return (
                            <ListItem button className={classes.nested} key={"target_"+i}>
                                <ListItemIcon>
                                    <StarBorder/>
                                </ListItemIcon>
                                <ListItemText primary={target.url}/>
                                <ListItemSecondaryAction>
                                    <IconButton edge="end" aria-label="delete">
                                        <DeleteIcon />
                                    </IconButton>
                                </ListItemSecondaryAction>
                            </ListItem>
                        )
                    })}
                </List>
            </Collapse>
            }
            <ListItem button onClick={() => handleClick('slack')}>
                <ListItemIcon>
                    <InboxIcon />
                </ListItemIcon>
                <ListItemText primary={"Slack URLS ("+item.slack_urls.length+")"} />
                {open.get('slack') ? <ExpandLess /> : <ExpandMore />}
            </ListItem>
            {item.slack_urls.length > 0 &&
            <Collapse in={open.get('slack')} timeout="auto" unmountOnExit>
                <List component="div" disablePadding>
                    {item.slack_urls.map((url, i) => {
                        return (
                            <ListItem button className={classes.nested} key={"slack_url_"+i}>
                                <ListItemIcon>
                                    <StarBorder/>
                                </ListItemIcon>
                                <ListItemText primary={url}/>
                                <ListItemSecondaryAction>
                                    <IconButton edge="end" aria-label="delete">
                                        <DeleteIcon />
                                    </IconButton>
                                </ListItemSecondaryAction>
                            </ListItem>
                        )
                    })}
                </List>
            </Collapse>
            }
        </List>
    );
}
