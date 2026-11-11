#!/bin/bash

exitcode=0

# The resources are described in internal/fwprovider (the device tag resource is written by hand).
resources="$(grep -ho 'TypeName: *"netbox_[a-z0-9_]*"' internal/fwprovider/resources_*.go | sed 's/.*"\(.*\)"/\1/' | sort -u) netbox_device_tag"

for resource in ${resources}; do
    ls -1 examples/resources/"$resource"/*.tf >/dev/null 2>&1
    # ls has exitcode 2 if no files are found
    if [ "$?" = "2" ]; then
        echo "Resource $resource has no example"
        exitcode=1
    fi
done

# Code for data source examples. Currently not used.
#datasources="$(grep -ho 'TypeName: *"netbox_[a-z0-9_]*"' internal/fwprovider/datasources_*.go | sed 's/.*"\(.*\)"/\1/' | sort -u)"
#
#for datasource in ${datasources}; do
#    ls -1 examples/data-sources/"$datasource"/*.tf >/dev/null 2>&1
#    if [ "$?" = "2" ]; then
#        echo "Data source $datasource has no example"
#        exitcode=1
#    fi
#done

exit $exitcode
