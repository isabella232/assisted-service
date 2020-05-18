// Code generated by go-swagger; DO NOT EDIT.

package installer

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

// NewDisableHostParams creates a new DisableHostParams object
// with the default values initialized.
func NewDisableHostParams() *DisableHostParams {
	var ()
	return &DisableHostParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewDisableHostParamsWithTimeout creates a new DisableHostParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewDisableHostParamsWithTimeout(timeout time.Duration) *DisableHostParams {
	var ()
	return &DisableHostParams{

		timeout: timeout,
	}
}

// NewDisableHostParamsWithContext creates a new DisableHostParams object
// with the default values initialized, and the ability to set a context for a request
func NewDisableHostParamsWithContext(ctx context.Context) *DisableHostParams {
	var ()
	return &DisableHostParams{

		Context: ctx,
	}
}

// NewDisableHostParamsWithHTTPClient creates a new DisableHostParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewDisableHostParamsWithHTTPClient(client *http.Client) *DisableHostParams {
	var ()
	return &DisableHostParams{
		HTTPClient: client,
	}
}

/*DisableHostParams contains all the parameters to send to the API endpoint
for the disable host operation typically these are written to a http.Request
*/
type DisableHostParams struct {

	/*ClusterID*/
	ClusterID strfmt.UUID
	/*HostID*/
	HostID strfmt.UUID

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the disable host params
func (o *DisableHostParams) WithTimeout(timeout time.Duration) *DisableHostParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the disable host params
func (o *DisableHostParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the disable host params
func (o *DisableHostParams) WithContext(ctx context.Context) *DisableHostParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the disable host params
func (o *DisableHostParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the disable host params
func (o *DisableHostParams) WithHTTPClient(client *http.Client) *DisableHostParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the disable host params
func (o *DisableHostParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithClusterID adds the clusterID to the disable host params
func (o *DisableHostParams) WithClusterID(clusterID strfmt.UUID) *DisableHostParams {
	o.SetClusterID(clusterID)
	return o
}

// SetClusterID adds the clusterId to the disable host params
func (o *DisableHostParams) SetClusterID(clusterID strfmt.UUID) {
	o.ClusterID = clusterID
}

// WithHostID adds the hostID to the disable host params
func (o *DisableHostParams) WithHostID(hostID strfmt.UUID) *DisableHostParams {
	o.SetHostID(hostID)
	return o
}

// SetHostID adds the hostId to the disable host params
func (o *DisableHostParams) SetHostID(hostID strfmt.UUID) {
	o.HostID = hostID
}

// WriteToRequest writes these params to a swagger request
func (o *DisableHostParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param cluster_id
	if err := r.SetPathParam("cluster_id", o.ClusterID.String()); err != nil {
		return err
	}

	// path param host_id
	if err := r.SetPathParam("host_id", o.HostID.String()); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}